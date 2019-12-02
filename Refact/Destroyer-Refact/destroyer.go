package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"
)

const (
	_respuesta = iota
	_registro
	_notificacion
	_check
)

type Mensaje struct {
	Codigo   int
	DirProp  string
	DirOtros []string
	Num      int
}

var MyDir string
var codigos map[int]string
var chDirOtros chan []string

func main() {
	fmt.Print("Ingrese la direccion del destructor (IP:Port): ")
	fmt.Scanln(&MyDir)
	generarCodigoDestructor()
	conectarAlSiguiente()
	server()
}

func generarCodigoDestructor() {
	rand.Seed(time.Now().UTC().UnixNano())
	n := rand.Intn(10)
	codigos = make(map[int]string)
	for i := 0; i < n; i++ {
		codigos[rand.Intn(100)] = ""
	}
	fmt.Println(codigos)
}
func conectarAlSiguiente() {
	chDirOtros = make(chan []string)
	var remote string
	fmt.Print("Ingrese la direccion del siguiente destructor (IP:Port): ")
	fmt.Scanln(&remote)
	if remote != "" {
		go sendMensajeOtroDestructor(remote, Mensaje{_registro, MyDir, []string{}, 0},
			func(conn net.Conn) {
				var mensaje Mensaje
				dec := json.NewDecoder(conn)
				dec.Decode(&mensaje)
				fmt.Println("Resp", mensaje)
				chDirOtros <- append(mensaje.DirOtros, remote)
			})
	} else {
		go func() { chDirOtros <- make([]string, 0, 10) }()
	}
}
func server() {
	if ln, err := net.Listen("tcp", MyDir); err != nil {
		log.Panicln(err.Error())
	} else {
		defer ln.Close()
		fmt.Println(MyDir, "escuchando")
		for {
			if conn, err := ln.Accept(); err != nil {
				log.Panicln(err.Error())
			} else {
				go handle(conn)
			}
		}
	}
}

func sendMensajeOtroDestructor(remoteAddr string, mensaje Mensaje, resp func(c net.Conn)) {
	if conn, err := net.Dial("tcp", remoteAddr); err != nil {
		log.Println(err.Error())
	} else {
		defer conn.Close()
		enc := json.NewEncoder(conn)
		fmt.Println("Mandando", mensaje, "to", remoteAddr)
		enc.Encode(&mensaje)
		if resp != nil {
			resp(conn)
		}
	}
}

func send(remoteAddr string, mensaje Mensaje) {
	sendMensajeOtroDestructor(remoteAddr, mensaje, nil)
}

func handle(conn net.Conn) {
	defer conn.Close()
	fmt.Println(conn.RemoteAddr(), "accepted")
	var mensaje Mensaje
	dec := json.NewDecoder(conn)
	if err := dec.Decode(&mensaje); err != nil {
		log.Println(err.Error())
	} else {
		fmt.Println("Got", mensaje)
		switch mensaje.Codigo {
		case _registro:
			registrar(conn, mensaje)
		case _notificacion:
			notificar(mensaje)
		case _check:
			check(conn, mensaje)
		}
	}
}

func registrar(conn net.Conn, mensaje Mensaje) {
	addrs := <-chDirOtros
	enc := json.NewEncoder(conn)
	enc.Encode(&Mensaje{_respuesta, MyDir, addrs, 0})
	for _, addr := range addrs {
		send(addr, Mensaje{_notificacion, mensaje.DirProp, []string{}, 0})
	}
	go addAddr(addrs, mensaje.DirProp)
}
func notificar(mensaje Mensaje) {
	fmt.Println(chDirOtros)
	addrs := <-chDirOtros
	go addAddr(addrs, mensaje.DirProp)
}
func check(conn net.Conn, mensaje Mensaje) {
	enc := json.NewEncoder(conn)
	num := -1
	if _, ok := codigos[mensaje.Num]; ok {
		num = mensaje.Num
	}
	enc.Encode(&Mensaje{_respuesta, MyDir, []string{}, num})
}

func addAddr(addrs []string, addr string) {
	chDirOtros <- append(addrs, addr)
	fmt.Println(addr, "added")
}
