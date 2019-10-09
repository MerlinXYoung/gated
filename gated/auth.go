package main

type IAuth interface{
	doAuth(opend string, openkey string)(uint64, error)
}
