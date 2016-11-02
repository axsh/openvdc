package util

import (
        "log"
        "os"
)

func SetupLog(logpath string, filename string, prefix string){

        if _, err := os.Stat(logpath); os.IsNotExist(err){
                os.Mkdir(logpath, os.ModePerm)
        }

        if _, err := os.Stat(logpath + filename); os.IsNotExist(err){
                _, err := os.Create(logpath + filename)
                if err != nil{
                        log.Println("Error creating log file: ", err)
                }
        }

        vdclog, err := os.OpenFile(logpath + filename, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
        if err != nil {
           log.Println("ERROR: Couldn't open log file", err)
        }

        log.SetOutput(vdclog)
        log.SetPrefix(prefix)
}
