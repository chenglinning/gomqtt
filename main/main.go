package server
import (
	"net"
	"errors"
	"regexp"
	"sync"
	"time"
	"github.com/gomqtt/mqttp"
)

type Server interface {

}

func process(conn net.Conn){
	defer conn.Close()
	for {
	   var buf [128]byte
	   n ,err := conn.Read(buf[:])
	   if err != nil {
		  fmt.Println("Read from tcp server failed,err:",err)
		  break
	   }
	   data := string(buf[:n])
	   fmt.Printf("Recived from client,data:%s\n",data)
	}
 }
 
 func main() {
	// close TCP service port
	listener,err := net.Listen("tcp","0.0.0.0:9090")
	if err != nil {
	   fmt.Println("Listen tcp server failed,err:",err)
	   return
	}
 
	for{
	   // create socket connection
	   conn,err := listener.Accept()
	   if err != nil {
		  fmt.Println("Listen.Accept failed,err:",err)
		  continue
	   }
 
	   // handle connection
	   go process(conn)
	}
 }
  