package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	cmds "github.com/ipfs/go-ipfs-cmds"
	"github.com/ipfs/go-ipfs/core/commands/cmdenv"
	"github.com/ipfs/go-ipfs/core/mine/crypto"
	"github.com/ipfs/go-ipfs/core/mine/types"
	"os"
	"text/tabwriter"
)

type Account struct {
	Address     string
	BnbBalance  string
	AntzBalance string
}

type WalletAccount struct {
	Accounts       []Account
	DefaultAccount string
}

type stringOutput struct {
	Str string
}

// WalletCmd is the 'ant wallet' command
var WalletCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Interact with the wallet.",
		ShortDescription: `Interact with the wallet.`,
	},
	Options: []cmds.Option{},
	Subcommands: map[string]*cmds.Command{
		"ls":         AddressListCmd,
		"new":        AddressNewCmd,
		"delete":     AddressDeleteCmd,
		"import":     AddressImportCmd,
		"export":     AddressExportCmd,
		"default":    AddressGetDefaultCmd,
		"setdefault": AddressSetDefaultCmd,
	},
}

var AddressListCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Get wallet address list",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {

		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}
		list, err := api.Wallet().List(req.Context)
		if err != nil {
			return err
		}
		var accounts []Account
		for _, key := range list {
			addr := crypto.EthereumAddress(key.PublicKey)
			bnb, err := api.Chain().GetBnbBalanceOf(req.Context, addr)
			if err != nil {
				continue
			}
			antz, err := api.Chain().GetAntzBalanceOf(req.Context, addr)
			if err != nil {
				continue
			}
			account := Account{
				Address:     addr,
				BnbBalance:  types.NBNFromRawString(bnb.String()).String(),
				AntzBalance: types.AntzFromRawString(antz.String()).String(),
			}
			accounts = append(accounts, account)
		}
		addr, err := api.Wallet().GetDefaultAddress(req.Context)
		if err != nil {
			return err
		}
		return res.Emit(
			&WalletAccount{
				Accounts:       accounts,
				DefaultAccount: crypto.EthereumAddress(addr.PublicKey),
			})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			v, _ := res.Next()
			accounts, ok := v.(*WalletAccount)
			if !ok {
				data, _ := json.MarshalIndent(accounts, " ", " ")
				fmt.Fprintf(os.Stdout, "%s\n", string(data))
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 15, 4, 1, ' ', tabwriter.AlignRight)

			for _, account := range accounts.Accounts {
				if account.Address == accounts.DefaultAccount {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", account.Address, account.BnbBalance, account.AntzBalance, "default")
				} else {
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n", account.Address, account.BnbBalance, account.AntzBalance, "")
				}
			}
			w.Flush()

			return nil
		},
	},
	Type: WalletAccount{},
}

var AddressNewCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "New wallet address",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {

		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}
		key, err := api.Wallet().NewAddress(req.Context)
		if err != nil {
			return err
		}
		addr := crypto.EthereumAddress(key.PublicKey)
		return res.Emit(&stringOutput{Str: addr})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			addr, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", addr.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}

var AddressImportCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Import wallet address",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("privatekey", true, false, "private key hex"),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		key, err := api.Wallet().Import(req.Context, req.Arguments[0])
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		addr := crypto.EthereumAddress(key.PublicKey)
		return res.Emit(&stringOutput{Str: addr})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			addr, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", addr.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}

var AddressDeleteCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Delete wallet address",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("address", true, false, "wallet address"),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		err = api.Wallet().Delete(req.Context, req.Arguments[0])
		if err != nil {
			res.Emit(&stringOutput{Str: err.Error()})
			return err
		}
		return res.Emit(&stringOutput{Str: "Ok"})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			addr, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", addr.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}

var AddressSetDefaultCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Set default wallet address",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("address", true, false, "wallet address"),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			res.Emit(res.Emit(&stringOutput{Str: err.Error()}))
			return err
		}
		err = api.Wallet().SetDefaultAddress(req.Context, req.Arguments[0])
		if err != nil {
			res.Emit(res.Emit(&stringOutput{Str: err.Error()}))
			return err
		}
		return res.Emit(&stringOutput{Str: "Ok"})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			addr, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", addr.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}

var AddressGetDefaultCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Get default wallet address",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{},
	Options:   []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			res.Emit(res.Emit(&stringOutput{Str: err.Error()}))
			return err
		}
		key, err := api.Wallet().GetDefaultAddress(req.Context)
		if err != nil {
			res.Emit(res.Emit(&stringOutput{Str: err.Error()}))
			return err
		}
		addr := crypto.EthereumAddress(key.PublicKey)
		return res.Emit(&stringOutput{Str: addr})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			addr, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", addr.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}

var AddressExportCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Export private key",
		ShortDescription: ``,
	},
	Arguments: []cmds.Argument{
		cmds.StringArg("address", true, false, "wallet address"),
	},
	Options: []cmds.Option{},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			res.Emit(res.Emit(&stringOutput{Str: err.Error()}))
			return err
		}
		key, err := api.Wallet().Get(req.Context, req.Arguments[0])
		if err != nil {
			res.Emit(res.Emit(&stringOutput{Str: err.Error()}))
			return err
		}
		prvk := hex.EncodeToString(crypto.EncodeSecp256k1PrivateKey(key))
		return res.Emit(&stringOutput{Str: prvk})
	},
	PostRun: cmds.PostRunMap{
		cmds.CLI: func(res cmds.Response, re cmds.ResponseEmitter) error {
			addr, _ := res.Next()
			fmt.Fprintf(os.Stdout, "%s\n", addr.(*stringOutput).Str)
			return nil
		},
	},
	Type: stringOutput{},
}
