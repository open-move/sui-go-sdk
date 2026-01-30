package transaction

import (
	bcs "github.com/iotaledger/bcs-go"
	"github.com/open-move/sui-go-sdk/types"
	"github.com/open-move/sui-go-sdk/typetag"
	"github.com/open-move/sui-go-sdk/utils"
)

// MoveCall represents a Move call command.
type MoveCall struct {
	Target        string
	Package       string
	Module        string
	Function      string
	TypeArguments []string
	Arguments     []Argument
}

// MakeMoveVecInput represents the input for a MakeMoveVec command.
type MakeMoveVecInput struct {
	Type     *string
	Elements []Argument
}

type PublishInput struct {
	Modules      [][]byte
	Dependencies []string
}

// UpgradeInput represents the input for an Upgrade command.
type UpgradeInput struct {
	Modules      [][]byte
	Dependencies []string
	Package      string
	Ticket       Argument
}

func (m MoveCall) toProgrammableMoveCall() (ProgrammableMoveCall, error) {
	pkg := m.Package
	mod := m.Module
	fn := m.Function
	if m.Target != "" {
		parsedPkg, parsedMod, parsedFn, err := utils.ParseMoveCallTarget(m.Target)
		if err != nil {
			return ProgrammableMoveCall{}, err
		}

		pkg = parsedPkg
		mod = parsedMod
		fn = parsedFn
	}

	if pkg == "" || mod == "" || fn == "" {
		return ProgrammableMoveCall{}, ErrMissingMoveCallTarget
	}

	address, err := utils.ParseAddress(pkg)
	if err != nil {
		return ProgrammableMoveCall{}, err
	}

	parsedTypeArgs := make([]typetag.TypeTag, len(m.TypeArguments))
	for i, arg := range m.TypeArguments {
		parsed, err := utils.ParseTypeTag(arg)
		if err != nil {
			return ProgrammableMoveCall{}, err
		}
		parsedTypeArgs[i] = parsed
	}

	return ProgrammableMoveCall{
		Package:       address,
		Module:        mod,
		Function:      fn,
		TypeArguments: parsedTypeArgs,
		Arguments:     append([]Argument(nil), m.Arguments...),
	}, nil
}

func (m MakeMoveVecInput) toCommand() (MakeMoveVec, error) {
	var tag *typetag.TypeTag
	if m.Type != nil {
		parsed, err := utils.ParseTypeTag(*m.Type)
		if err != nil {
			return MakeMoveVec{}, err
		}
		tag = &parsed
	}
	return MakeMoveVec{
		Type:     optionTypeTag(tag),
		Elements: append([]Argument(nil), m.Elements...),
	}, nil
}

func (p PublishInput) toCommand() (Publish, error) {
	deps, err := parseAddresses(p.Dependencies)
	if err != nil {
		return Publish{}, err
	}

	return Publish{
		Modules:      cloneModules(p.Modules),
		Dependencies: deps,
	}, nil
}

func (u UpgradeInput) toCommand() (Upgrade, error) {
	deps, err := parseAddresses(u.Dependencies)
	if err != nil {
		return Upgrade{}, err
	}

	pkg, err := utils.ParseAddress(u.Package)
	if err != nil {
		return Upgrade{}, err
	}

	return Upgrade{
		Modules:      cloneModules(u.Modules),
		Dependencies: deps,
		Package:      pkg,
		Ticket:       u.Ticket,
	}, nil
}

func optionTypeTag(tag *typetag.TypeTag) bcs.Option[typetag.TypeTag] {
	if tag == nil {
		return bcs.Option[typetag.TypeTag]{None: true}
	}

	return bcs.Option[typetag.TypeTag]{Some: *tag}
}

func parseAddresses(addresses []string) ([]types.Address, error) {
	if len(addresses) == 0 {
		return nil, nil
	}

	parsed := make([]types.Address, len(addresses))
	for i, addr := range addresses {
		parsedAddr, err := utils.ParseAddress(addr)
		if err != nil {
			return nil, err
		}

		parsed[i] = parsedAddr
	}

	return parsed, nil
}

func cloneModules(modules [][]byte) [][]byte {
	if len(modules) == 0 {
		return nil
	}

	cloned := make([][]byte, len(modules))
	for i, module := range modules {
		cloned[i] = append([]byte(nil), module...)
	}

	return cloned
}
