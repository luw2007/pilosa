package transport

import (
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"pilosa/config"
	"pilosa/core"
	"pilosa/db"

	"time"
	"tux21b.org/v1/gocql/uuid"
)

type envelope struct {
	message *db.Message
	host    *uuid.UUID
}

type connection struct {
	transport *TcpTransport
	inbox     chan *db.Message
	outbox    chan *db.Message
	conn      *net.Conn
	process   *uuid.UUID
}

type newconnection struct {
	id         *uuid.UUID
	connection *connection
}

func init() {
	gob.Register(uuid.UUID{})
}

func (self *connection) manage() {
BeginManageConnection:
	for {
		log.Println("manage", self)
		if self.conn == nil {
			process, err := self.transport.service.ProcessMap.GetProcess(self.process)
			if err != nil {
				log.Println("transport/tcp: error getting process, retrying in 2 seconds... ", self.process, err)
				time.Sleep(2 * time.Second)
				continue
			}
			host_string := fmt.Sprintf("%s:%d", process.Host(), process.PortTcp())
			conn, err := net.Dial("tcp", host_string)
			if err != nil {
				log.Println("transport/tcp: error dialing: ", host_string, " Retrying in 2 seconds...")
				time.Sleep(2 * time.Second)
				continue
			}
			self.conn = &conn
			go func() {
				self.outbox <- &db.Message{self.transport.service.Id}
			}()
		}
		encoder := gob.NewEncoder(*self.conn)
		decoder := gob.NewDecoder(*self.conn)
		var exit = make(chan int)
		go func() {
			for {
				var mess *db.Message
				err := decoder.Decode(&mess)
				if err != nil {
					log.Println("transport/tcp: error decoding message: ", err.Error())
					exit <- 1
					return
				}
				self.inbox <- mess
			}
		}()
		for {
			select {
			case message := <-self.outbox:
				err := encoder.Encode(message)
				if err != nil {
					log.Println(err.Error())
					return
				}
			case message := <-self.inbox:
				identifier, ok := message.Data.(uuid.UUID)
				if ok {
					// message is connection registration; bypass inbox and register
					self.process = &identifier
					self.transport.reg <- &newconnection{&identifier, self}
				} else {
					self.transport.inbox <- message
				}
			case <-exit:
				if self.process != nil {
					self.conn = nil
					continue BeginManageConnection
				} else {
					return
				}
			}
		}
	}
}

type TcpTransport struct {
	service     *core.Service
	port        int
	inbox       chan *db.Message
	outbox      chan envelope
	connections map[uuid.UUID]*connection
	reg         chan *newconnection
}

func (self *TcpTransport) Run() {
	log.Println("Initializing TCP transport")
	go self.listen()
	for {
		select {
		case env := <-self.outbox:
			con, ok := self.connections[*(env.host)]
			if !ok {
				con = &connection{self, make(chan *db.Message, 100), make(chan *db.Message, 100), nil, env.host}
				go con.manage()
				self.connections[*env.host] = con
			}
			con.outbox <- env.message
		case nc := <-self.reg:
			self.connections[*nc.id] = nc.connection
		}
	}
}

func (self *TcpTransport) listen() {
	port_string := fmt.Sprintf(":%d", self.port)
	l, e := net.Listen("tcp", port_string)
	if e != nil {
		log.Fatal("Cannot bind to port! ", self.port)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting, trying again in 2 sec... ", err)
			time.Sleep(2 * time.Second)
			continue
		}
		go self.manage(&conn)
	}
}

func (self *TcpTransport) manage(conn *net.Conn) {
	con := &connection{self, make(chan *db.Message, 100), make(chan *db.Message, 100), conn, nil}
	con.manage()
}

func (self *TcpTransport) Close() {
	log.Println("Shutting down TCP transport")
}

func (self *TcpTransport) Send(message *db.Message, host *uuid.UUID) {
	self.outbox <- envelope{message, host}
}

func (self *TcpTransport) Receive() *db.Message {
	return <-self.inbox
}

func (self *TcpTransport) Push(message *db.Message) {
	self.inbox <- message
}

func NewTcpTransport(service *core.Service) *TcpTransport {
	return &TcpTransport{service, config.GetInt("port_tcp"), make(chan *db.Message, 100), make(chan envelope, 100), make(map[uuid.UUID]*connection), make(chan *newconnection)}
}