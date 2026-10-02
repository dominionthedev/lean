package lean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dominionthedev/lean/internal/backup"
	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var profileCmd=&cobra.Command{Use:"profile",Short:"Manage environment profiles"}

var profileDeleteCmd=&cobra.Command{
	Use:"delete NAME",Aliases:[]string{"rm"},Short:"Delete an environment profile",Args:cobra.ExactArgs(1),
	Run:func(cmd *cobra.Command,args []string){
		name:=args[0];engine,err:=core.NewEngine();if err!=nil{fmt.Println(ui.Fail("Not initialized. Run lean init first."));return};_ = engine.ScanDisk()
		if !engine.ProfileExists(name){fmt.Println(ui.Fail(fmt.Sprintf("Profile '%s' not found.",name)));return}
		if engine.State.Current==name{fmt.Println(ui.Fail(fmt.Sprintf("Cannot delete active profile '%s'. Apply another profile first.",name)));return}
		path:=env.ProfilePath(name);if _,err:=os.Stat(path);err!=nil{fmt.Println(ui.Warn(fmt.Sprintf("Profile '%s' no longer exists. Removing it from Lean.",name)));_ = engine.DeleteProfile(name);return}
		// Refuse deletion while a child depends on this profile.
		entries,_:=os.ReadDir(filepath.Join(".lean","profiles"))
		for _,entry:=range entries{if entry.IsDir()||!strings.HasSuffix(entry.Name(),".env"){continue};child:=strings.TrimSuffix(entry.Name(),".env");if child==name{continue};f,err:=env.Parse(filepath.Join(".lean","profiles",entry.Name()));if err==nil&&f.Extends()==name{fmt.Println(ui.Fail(fmt.Sprintf("Cannot delete '%s': profile '%s' extends it.",name,child)));return}}
		if err:=backup.SnapshotProfile(name,path);err!=nil{fmt.Println(ui.Warn("Could not back up profile: "+err.Error()))}
		if err:=os.RemoveAll(filepath.Join(".lean","secrets",name));err!=nil{fmt.Println(ui.Fail("Failed to remove profile secrets: "+err.Error()));return}
		if err:=os.Remove(path);err!=nil{fmt.Println(ui.Fail("Failed to delete profile: "+err.Error()));return}
		if err:=engine.DeleteProfile(name);err!=nil{fmt.Println(ui.Fail("Failed to update profile state: "+err.Error()));return}
		fmt.Println(ui.Ok(fmt.Sprintf("Profile '%s' deleted.",name)))
	},
}
func init(){profileCmd.AddCommand(profileDeleteCmd)}
