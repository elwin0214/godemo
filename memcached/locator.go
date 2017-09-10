package memcached

import (
	"hash/crc32"
	"math/rand"
	"time"
	"sync/atomic"
	. "github.com/elwin0214/gomemcached/util"
	"github.com/golang/glog"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type SessionLocator interface {
	getSessionByKey(key string) *Session

	updateSessions(sessionMap map[string]*List)
}

type ArraySessionLocator struct {
	addressList []*AddressInfo
	value       atomic.Value
	//sessions []*List
}

func newArraySessionLocator(addressList []*AddressInfo) SessionLocator {
	s := new(ArraySessionLocator)
	s.addressList = make([]*AddressInfo, 0, 2 * len(addressList))
	for _, addressInfo := range addressList {
		for i := 0; i < addressInfo.Weight; i++ {
			s.addressList = append(s.addressList, addressInfo)
		}
	}
	return s
}

func (s *ArraySessionLocator) getSessionByKey(key string) *Session {

	num := crc32.ChecksumIEEE([]byte(key))
	index := int(num) % len(s.addressList)
	addressInfo := s.addressList[index]
	sessions, _ := s.value.Load().(map[string]*List)
	sessionList := sessions[addressInfo.Address]
	return s.getRandomSession(sessionList)
}

func (s *ArraySessionLocator) getRandomSession(list *List) *Session {
	if nil == list || 0 == list.Len() {
		glog.Infof("[getRandomSession] size of sessions is 0")
		return nil
	}
	index := rand.Intn(list.Len())
	session, _ := (*list)[index].(*Session)
	return session
}

func (s *ArraySessionLocator) updateSessions(sessionMap map[string]*List) {
	newSessionMap := make(map[string]*List,  len(sessionMap))

	for _, sessionList := range sessionMap {
		s := (*sessionList)[0]
		session, _ := s.(*Session)
		addressInfo := session.addressInfo
		newSessionList, _ := sessionList.Clone()
		newSessionMap[addressInfo.Address] = &newSessionList

	}
	s.value.Store(newSessionMap)
}
