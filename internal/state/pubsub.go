package state

import (
	"context"
	"fmt"

	"github.com/0x222fe/codecrafters-redis-go/internal/connection"
	"github.com/0x222fe/codecrafters-redis-go/internal/utils/resputil"
	"github.com/google/uuid"
)

const (
	subChanBufSize = 16
)

type Subscriber struct {
	Conn     *connection.Connection
	Channels map[string]struct{}
	MsgChan  chan PubSubMsg
	Ctx      context.Context
	Cancel   context.CancelFunc
}

type PubSubMsg struct {
	Channel string
	Payload []byte
}

func (s *AppState) AddSubscriber(conn *connection.Connection, channel string) *Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscribers[conn.ID]
	if !ok {
		ctx, cancel := context.WithCancel(context.Background())
		sub = &Subscriber{
			Conn:     conn,
			Ctx:      ctx,
			Channels: make(map[string]struct{}),
			Cancel:   cancel,
			MsgChan:  make(chan PubSubMsg, subChanBufSize),
		}
		s.subscribers[conn.ID] = sub
		go func() {
			for {
				select {
				case msg := <-sub.MsgChan:
					conn.WriteResp(resputil.BulkStringsToRESPArray([]string{
						"message",
						msg.Channel,
						string(msg.Payload),
					}))
				case <-sub.Ctx.Done():
					return
				}
			}
		}()
	}
	sub.Channels[channel] = struct{}{}

	chanMap, ok := s.channelSubs[channel]
	if !ok {
		chanMap = make(map[uuid.UUID]*Subscriber)
		s.channelSubs[channel] = chanMap
	}

	chanMap[conn.ID] = sub

	fmt.Printf("Subscriber connected: %s\n", conn.RemoteAddr().String())

	return sub
}

func (s *AppState) UnsubChannel(id uuid.UUID, channel string) *Subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()

	sub, ok := s.subscribers[id]
	if !ok {
		return nil
	}

	chanMap, ok := s.channelSubs[channel]
	if !ok {
		return sub
	}

	_, ok = chanMap[id]
	if !ok {
		return sub
	}

	delete(chanMap, id)
	delete(sub.Channels, channel)

	return sub
}

func (s *AppState) RemoveSubscriber(id uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sub, exists := s.subscribers[id]; exists {
		sub.Cancel()

		for channel := range sub.Channels {
			delete(s.channelSubs[channel], id)
		}

		delete(s.subscribers, id)
		fmt.Printf("Subscriber disconnected: %s\n", sub.Conn.RemoteAddr().String())
	}
}

func (s *AppState) Publish(channel string, payload []byte) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sent := 0
	chanMap, ok := s.channelSubs[channel]
	if !ok {
		return sent
	}

	for _, sub := range chanMap {
		select {
		case sub.MsgChan <- PubSubMsg{Channel: channel, Payload: payload}:
			sent++
		default:
		}
	}

	return sent
}
