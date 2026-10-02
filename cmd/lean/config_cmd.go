package lean

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dominionthedev/lean/internal/config"
	"github.com/dominionthedev/lean/internal/ui"
	"github.com/spf13/cobra"
)

var (
	cfgDefaultProfile string
	cfgSchemaPath string
	cfgSecretsBackend string
	cfgSecretsRecipient string
	cfgSecretsIdentity string
	cfgMasterKeyFile string
	cfgGlobal bool
	cfgInteractive bool
)

var configCmd=&cobra.Command{
	Use:"config",Short:"View or initialize lean configuration",
	Run:func(cmd *cobra.Command,args []string){
		cfg,err:=config.Load("");if err!=nil{fmt.Println(ui.Fail("Failed to load config: "+err.Error()));return}
		fmt.Printf("%s lean.toml\n\n",ui.Bolt())
		if !config.Exists(""){fmt.Println(ui.Faint("  (using defaults — run lean config init to write local config)"));fmt.Println()}
		fmt.Printf("  %s %d\n",ui.Faint("version"),cfg.Lean.Version)
		fmt.Printf("  %s %s\n",ui.Faint("default_profile"),orDash(cfg.Lean.DefaultProfile))
		fmt.Printf("  %s %s\n",ui.Faint("schema.path"),cfg.Schema.Path)
		fmt.Printf("  %s %s\n",ui.Faint("secrets.backend"),cfg.Secrets.Backend)
		if cfg.Secrets.Recipient!=""{fmt.Printf("  %s %s\n",ui.Faint("secrets.recipient"),cfg.Secrets.Recipient)}
		if cfg.Secrets.Identity!=""{fmt.Printf("  %s %s\n",ui.Faint("secrets.identity"),cfg.Secrets.Identity)}
		fmt.Printf("  %s %s\n",ui.Faint("secrets.master_key_file"),orDash(cfg.Secrets.MasterKeyFile))
	},
}

var configInitCmd=&cobra.Command{
	Use:"init",Short:"Create a local or global lean.toml",
	Run:func(cmd *cobra.Command,args []string){
		global:=cfgGlobal
		if cfgInteractive||(!cmd.Flags().Changed("global")&& !cfgInteractive){
			var scope string
			if cfgGlobal{scope="global"}else{scope="local"}
			form:=huh.NewForm(huh.NewGroup(
				huh.NewSelect[string]().Title("Configuration scope").Options(
					huh.NewOption("Local project (.lean/lean.toml)","local"),
					huh.NewOption("Global user (~/.lean/lean.toml)","global"),
				).Value(&scope),
			))
			if err:=form.Run();err!=nil{fmt.Println(ui.Fail("Interrupted."));return}
			global=scope=="global"
			cfgInteractive=true
		}
		path:=config.LocalPath()
		if global{var err error;path,err=config.GlobalPath();if err!=nil{fmt.Println(ui.Fail("Cannot determine home directory: "+err.Error()));return}}
		if config.Exists(path){fmt.Println(ui.Warn(path+" already exists."));return}
		cfg:=config.Default()
		if cfgInteractive{
			form:=huh.NewForm(huh.NewGroup(
				huh.NewInput().Title("Default profile (optional)").Value(&cfg.Lean.DefaultProfile),
				huh.NewInput().Title("Schema path").Value(&cfg.Schema.Path),
				huh.NewSelect[string]().Title("Secrets backend").Options(
					huh.NewOption("Local","local"),huh.NewOption("GPG","gpg"),huh.NewOption("age","age"),huh.NewOption("SSH","ssh"),
				).Value(&cfg.Secrets.Backend),
			))
			if err:=form.Run();err!=nil{fmt.Println(ui.Fail("Interrupted."));return}
		}
		if err:=config.Save(path,cfg);err!=nil{fmt.Println(ui.Fail("Failed to write config: "+err.Error()));return}
		fmt.Println(ui.Ok("Wrote "+path))
	},
}

var configSetCmd=&cobra.Command{
	Use:"set",Short:"Update lean.toml settings",
	Run:func(cmd *cobra.Command,args []string){
		cfg,err:=config.Load("");if err!=nil{fmt.Println(ui.Fail("Failed to load config: "+err.Error()));return}
		changed:=false
		if cmd.Flags().Changed("default-profile"){cfg.Lean.DefaultProfile=cfgDefaultProfile;changed=true}
		if cmd.Flags().Changed("schema-path"){cfg.Schema.Path=cfgSchemaPath;changed=true}
		if cmd.Flags().Changed("secrets-backend"){cfg.Secrets.Backend=cfgSecretsBackend;changed=true}
		if cmd.Flags().Changed("secrets-recipient"){cfg.Secrets.Recipient=cfgSecretsRecipient;changed=true}
		if cmd.Flags().Changed("secrets-identity"){cfg.Secrets.Identity=cfgSecretsIdentity;changed=true}
		if cmd.Flags().Changed("master-key-file"){cfg.Secrets.MasterKeyFile=cfgMasterKeyFile;changed=true}
		if !changed{fmt.Println(ui.Info("Nothing to change."));return}
		if err:=config.Save("",cfg);err!=nil{fmt.Println(ui.Fail("Failed to write config: "+err.Error()));return}
		fmt.Println(ui.Ok("lean.toml updated."))
	},
}
func orDash(s string)string{if s==""{return "—"};return s}
func init(){
	configInitCmd.Flags().BoolVarP(&cfgGlobal,"global","g",false,"Create global config in ~/.lean")
	configInitCmd.Flags().BoolVarP(&cfgInteractive,"interactive","i",false,"Create the config interactively")
	configSetCmd.Flags().StringVar(&cfgDefaultProfile,"default-profile","","Default profile name")
	configSetCmd.Flags().StringVar(&cfgSchemaPath,"schema-path","","Path to schema.toml")
	configSetCmd.Flags().StringVar(&cfgSecretsBackend,"secrets-backend","","Secrets backend: local | gpg | age | ssh")
	configSetCmd.Flags().StringVar(&cfgSecretsRecipient,"secrets-recipient","","GPG/age recipient or SSH .pub path")
	configSetCmd.Flags().StringVar(&cfgSecretsIdentity,"secrets-identity","","Age identity / SSH private key path")
	configSetCmd.Flags().StringVar(&cfgMasterKeyFile,"master-key-file","","Path to master key file")
	configCmd.AddCommand(configInitCmd);configCmd.AddCommand(configSetCmd)
}
