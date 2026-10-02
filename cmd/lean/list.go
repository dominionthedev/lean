package lean

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dominionthedev/lean/internal/core"
	"github.com/dominionthedev/lean/internal/env"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd=&cobra.Command{
	Use:"list",Short:"List all environment profiles",
	Run:func(cmd *cobra.Command,args []string){
		engine,err:=core.NewEngine();if err!=nil{fmt.Println(ui.Fail("Not initialized. Run lean init first."));return}
		prev:=engine.State.Current;_ = engine.ScanDisk()
		if prev!=""&&engine.State.Current==""{fmt.Println(ui.Warn(fmt.Sprintf("Active profile '%s' no longer exists. It has been removed from Lean.",prev)))}
		if len(engine.State.Profiles)==0{fmt.Println(ui.Info("No profiles yet. Run lean create to make one."));return}
		fmt.Println(ui.Bolt()+" "+ui.Bold.Render("Profiles"));fmt.Println()
		for _,profile:=range engine.State.Profiles{
			extends:=""
			if f,err:=env.Parse(env.ProfilePath(profile));err==nil{if parent:=f.Extends();parent!=""{extends=" "+ui.Faint("→ "+parent)}}
			if profile==engine.State.Current{fmt.Printf("  %s %s%s\n",ui.Success.Render("▶"),ui.Active.Render(profile)+" "+ui.Faint("(active)"),extends)}else{fmt.Printf("  %s %s%s\n",ui.Muted.Render("·"),profile,extends)}
		}
		_ = os.MkdirAll(filepath.Join(".lean","profiles"),0700)
		fmt.Println()
	},
}
