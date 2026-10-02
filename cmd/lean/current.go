package lean

import (
	"fmt"
	"os"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var currentCmd=&cobra.Command{
	Use:"current",Short:"Show the active environment profile",
	Run:func(cmd *cobra.Command,args []string){
		engine,err:=core.NewEngine();if err!=nil{fmt.Println(ui.Fail("Not initialized. Run lean init first."));return}
		prev:=engine.State.Current;_ = engine.ScanDisk()
		if prev!=""&&engine.State.Current==""{fmt.Println(ui.Warn(fmt.Sprintf("Active profile '%s' no longer exists. It has been removed from Lean.",prev)));return}
		if engine.State.Current==""{fmt.Println(ui.Info("No active profile. Run lean apply <profile> to set one."));return}
		if _,err:=os.Stat(env.ProfilePath(engine.State.Current));err!=nil{fmt.Println(ui.Warn(fmt.Sprintf("Active profile '%s' no longer exists. It has been removed from Lean.",engine.State.Current)));_ = engine.DeleteProfile(engine.State.Current);return}
		fmt.Printf("%s %s\n",ui.Bolt(),ui.Active.Render(engine.State.Current))
	},
}
