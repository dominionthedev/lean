package core

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Engine struct { State *State }

func NewEngine()(*Engine,error){state,err:=LoadState();if err!=nil{return nil,err};return &Engine{State:state},nil}

func Initialize()error{s:=&State{Initialized:true,Version:"1.0.0",Profiles:[]string{},Templates:[]string{}};return SaveState(s)}

func(e *Engine)AddProfile(name string)error{for _,p:=range e.State.Profiles{if p==name{return errors.New("profile already exists")}};e.State.Profiles=append(e.State.Profiles,name);return SaveState(e.State)}

func(e *Engine)DeleteProfile(name string)error{for i,p:=range e.State.Profiles{if p==name{e.State.Profiles=append(e.State.Profiles[:i],e.State.Profiles[i+1:]...);if e.State.Meta!=nil{delete(e.State.Meta,name)};if e.State.Current==name{e.State.Current=""};return SaveState(e.State)}};return errors.New("profile not found")}

func(e *Engine)AddTemplate(path string)error{for _,t:=range e.State.Templates{if t==path{return nil}};e.State.Templates=append(e.State.Templates,path);return SaveState(e.State)}
func(e *Engine)ProfileExists(name string)bool{for _,p:=range e.State.Profiles{if p==name{return true}};return false}
func(e *Engine)SetCurrent(name string)error{e.State.Current=name;return SaveState(e.State)}

func(e *Engine)ScanDisk()error{
	if err:=os.MkdirAll(filepath.Join(".lean","profiles"),0700);err!=nil{return err}
	known:=make(map[string]bool)
	for _,p:=range e.State.Profiles{known[p]=true}

	// Migrate legacy .env.<name> files into .lean/profiles/<name>.env.
	entries,err:=os.ReadDir(".");if err!=nil{return err}
	for _,entry:=range entries{
		name:=entry.Name()
		if !strings.HasPrefix(name,".env."){continue}
		profile:=strings.TrimPrefix(name,".env.")
		if profile=="tmp"||profile=="template"||profile=="example"{continue}
		dst:=filepath.Join(".lean","profiles",profile+".env")
		if _,err:=os.Stat(dst);os.IsNotExist(err){
			if data,err:=os.ReadFile(name);err==nil{if err:=os.WriteFile(dst,data,0600);err==nil{_ = os.Remove(name)}}
		}
		known[profile]=true
	}
	entries,err=os.ReadDir(filepath.Join(".lean","profiles"));if err!=nil{return err}
	disk:=make(map[string]bool)
	for _,entry:=range entries{
		if entry.IsDir()||!strings.HasSuffix(entry.Name(),".env"){continue}
		name:=strings.TrimSuffix(entry.Name(),".env");disk[name]=true
		if !known[name]{e.State.Profiles=append(e.State.Profiles,name);known[name]=true}
	}
	var kept []string
	for _,p:=range e.State.Profiles{if disk[p]{kept=append(kept,p)}else if e.State.Current==p{e.State.Current=""}}
	e.State.Profiles=kept
	return SaveState(e.State)
}

func(e *Engine)ScanTemplates()error{entries,err:=os.ReadDir(".");if err!=nil{return err};known:=make(map[string]bool);for _,t:=range e.State.Templates{known[t]=true};for _,entry:=range entries{name:=entry.Name();if (name==".env.template"||name==".env.example")&&!known[name]{e.State.Templates=append(e.State.Templates,name);known[name]=true}};return SaveState(e.State)}
func(e *Engine)SetMeta(name string,meta ProfileMeta)error{if e.State.Meta==nil{e.State.Meta=make(map[string]ProfileMeta)};e.State.Meta[name]=meta;return SaveState(e.State)}
func(e *Engine)GetMeta(name string)(ProfileMeta,bool){if e.State.Meta==nil{return ProfileMeta{},false};m,ok:=e.State.Meta[name];return m,ok}
