package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const backupDir = ".lean/backups"
const profileBackupDir = ".lean/backups/profiles"

func Snapshot(activeProfile string) error {
	data, err := os.ReadFile(".env")
	if err != nil { if os.IsNotExist(err){return nil}; return err }
	if err:=os.MkdirAll(backupDir,0755);err!=nil{return err}
	label:=activeProfile;if label==""{label="unknown"}
	path:=filepath.Join(backupDir,fmt.Sprintf("%s-%s.env",label,time.Now().Format("20060102-150405")))
	return os.WriteFile(path,data,0600)
}

func SnapshotProfile(profile,path string) error {
	data,err:=os.ReadFile(path);if err!=nil{return err}
	dir:=filepath.Join(profileBackupDir,profile)
	if err:=os.MkdirAll(dir,0700);err!=nil{return err}
	return os.WriteFile(filepath.Join(dir,time.Now().Format("20060102-150405.000000000")+".env"),data,0600)
}

func RestoreProfile(profile string) error {
	dir:=filepath.Join(profileBackupDir,profile)
	entries,err:=os.ReadDir(dir);if err!=nil{return err}
	var names []string
	for _,e:=range entries{if !e.IsDir()&&strings.HasSuffix(e.Name(),".env"){names=append(names,e.Name())}}
	if len(names)==0{return fmt.Errorf("no profile backup exists for %s",profile)}
	sort.Strings(names)
	data,err:=os.ReadFile(filepath.Join(dir,names[len(names)-1]));if err!=nil{return err}
	path:=filepath.Join(".lean/profiles",profile+".env")
	if err:=os.MkdirAll(filepath.Dir(path),0700);err!=nil{return err}
	return os.WriteFile(path,data,0600)
}

func ListProfileBackups(profile string)([]string,error){
	entries,err:=os.ReadDir(filepath.Join(profileBackupDir,profile))
	if err!=nil{if os.IsNotExist(err){return []string{},nil};return nil,err}
	var names []string
	for _,e:=range entries{if !e.IsDir()&&strings.HasSuffix(e.Name(),".env"){names=append(names,e.Name())}}
	sort.Sort(sort.Reverse(sort.StringSlice(names)));return names,nil
}

func NamedSnapshot(name string) error {
	name=sanitizeName(name);if name==""{return fmt.Errorf("snapshot name cannot be empty")}
	data,err:=os.ReadFile(".env");if err!=nil{if os.IsNotExist(err){return fmt.Errorf("no .env to snapshot")};return err}
	if err:=os.MkdirAll(backupDir,0755);err!=nil{return err}
	return os.WriteFile(filepath.Join(backupDir,name+".env"),data,0600)
}
func List()([]string,error){entries,err:=os.ReadDir(backupDir);if err!=nil{if os.IsNotExist(err){return []string{},nil};return nil,err};var names []string;for _,e:=range entries{if !e.IsDir()&&strings.HasSuffix(e.Name(),".env"){names=append(names,e.Name())}};sort.Sort(sort.Reverse(sort.StringSlice(names)));return names,nil}
func ListNamed()([]string,error){all,err:=List();if err!=nil{return nil,err};var named []string;for _,n:=range all{base:=strings.TrimSuffix(n,".env");if !isTimestamped(base){named=append(named,base)}};sort.Strings(named);return named,nil}
func Restore(name string)error{data,err:=os.ReadFile(resolvePath(name));if err!=nil{return err};tmp:=".env.tmp";if err:=os.WriteFile(tmp,data,0600);err!=nil{return err};return os.Rename(tmp,".env")}
func Exists(name string)bool{_,err:=os.Stat(resolvePath(name));return err==nil}
func Delete(name string)error{return os.Remove(resolvePath(name))}
func resolvePath(name string)string{if strings.HasSuffix(name,".env"){return filepath.Join(backupDir,name)};candidate:=filepath.Join(backupDir,name+".env");if _,err:=os.Stat(candidate);err==nil{return candidate};return filepath.Join(backupDir,name)}
func sanitizeName(name string)string{name=strings.TrimSpace(name);name=strings.ReplaceAll(name," ","-");name=strings.TrimSuffix(name,".env");name=strings.ReplaceAll(name,"/","-");name=strings.ReplaceAll(name,"\\","-");return name}
func isTimestamped(base string)bool{parts:=strings.Split(base,"-");if len(parts)<2{return false};date,timePart:=parts[len(parts)-2],parts[len(parts)-1];if len(date)!=8||len(timePart)!=6{return false};for _,c:=range date+timePart{if c<'0'||c>'9'{return false}};return true}
