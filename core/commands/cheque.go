package commands

import (
	"encoding/json"
	"fmt"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	iface "github.com/ipfs/interface-go-ipfs-core"
	"os"
)

type Cheques struct {
	List []iface.Cheque
}

// ChequeCmd is the 'ipfs cheque' command
var ChequeCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Interact with cheques.",
		ShortDescription: `Interact with cheques.`,
	},
	Options: []cmds.Option{},
	Subcommands: map[string]*cmds.Command{
		"ls":         ChequeListCmd,
		"get":        ChequeGetCmd,
		"cashout":    ChequeCashOutCmd,
		"cashoutall": ChequeCashOutAllCmd,
	},
}

var ChequeListCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Get cheque list",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {

		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}
		chequeList, err := api.Cheque().List(req.Context)
		if err != nil {
			return err
		}
		//fmt.Printf("cheques: %+v\n", chequeList)
		return res.Emit(Cheques{List: chequeList})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			cheques, _ := res.Next()
			data, _ := json.MarshalIndent(cheques, " ", " ")
			fmt.Fprintf(os.Stdout, "%s\n", string(data))
			return nil
		},
	},
	Type: Cheques{},
}

var ChequeGetCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Get cheque",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("chequebook", true, false, "chequebook contract address"),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}
		chequebook := req.Arguments[0]
		cheque, err := api.Cheque().Get(req.Context, chequebook)
		if err != nil {
			return err
		}
		//fmt.Printf("cheque: %v", cheque)
		return res.Emit(cheque)
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			v, _ := res.Next()
			data, _ := json.MarshalIndent(v, " ", " ")
			fmt.Fprintf(os.Stdout, "%s\n", string(data))
			return nil
		},
	},
	Type: iface.Cheque{},
}

var ChequeCashOutCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "CashOut cheques",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("chequebook", true, false, "chequebook contract address"),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		chequebook := req.Arguments[0]
		err = api.Cheque().CashOut(req.Context, chequebook)
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		return res.Emit(&stringOutput{Str: "Ok"})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			v, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", v.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}

var ChequeCashOutAllCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "CashOut all cheques",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		err = api.Cheque().CashOutAll(req.Context)
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		return res.Emit(&stringOutput{Str: "Ok"})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			v, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", v.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}
