package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

func main() {
	debug := flag.Bool("debug", false, "Verbose logging")
	port := flag.Int("port", 25000, "TCP listen port")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Listening on %s", addr)
	for i := 0; ; i++ {
		conn, err := lis.Accept()
		if err != nil {
			// TODO: check if error is temporary and retry with backoff
			log.Fatal(err)
		}
		if *debug {
			log.Printf("%d: %v <-> %v\n", i, conn.LocalAddr(), conn.RemoteAddr())
		}
		go func() {
			if err := copy(conn, conn); err != nil {
				log.Printf("copy error: %v", err)
			}
			conn.Close()
		}()
	}
}

func copy(dst io.Writer, src io.Reader) error {
	buf := getBufWriter(dst)
	defer putBufWriter(buf)
	if _, err := io.Copy(buf, src); err != nil && err != io.EOF {
		return err
	}
	return buf.Flush()
}

var bufPool = sync.Pool{
	New: func() interface{} { return bufio.NewWriter(nil) },
}

func getBufWriter(w io.Writer) *bufio.Writer {
	bw := bufPool.Get().(*bufio.Writer)
	bw.Reset(w)
	return bw
}

func putBufWriter(w *bufio.Writer) {
	w.Reset(nil)
	bufPool.Put(w)
}
