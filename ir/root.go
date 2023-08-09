package ir

import (
	ast_pb "github.com/txpull/protos/dist/go/ast"
	ir_pb "github.com/txpull/protos/dist/go/ir"
	"github.com/txpull/solgo/ast"
	"github.com/txpull/solgo/eip"
)

// RootSourceUnit represents the root of a Solidity contract's AST as an IR node.
type RootSourceUnit struct {
	unit              *ast.RootNode   `json:"-"`
	NodeType          ast_pb.NodeType `json:"node_type"`
	EntryContractId   int64           `json:"entry_contract_id"`
	EntryContractName string          `json:"entry_contract_name"`
	ContractsCount    int32           `json:"contracts_count"`
	ContractTypes     []string        `json:"contract_types"`
	Eips              []*EIP          `json:"eips"`
	Contracts         []*Contract     `json:"contracts"`
}

// GetAST returns the underlying AST node of the RootSourceUnit.
func (r *RootSourceUnit) GetAST() *ast.RootNode {
	return r.unit
}

// GetNodeType returns the type of the node in the AST.
func (r *RootSourceUnit) GetNodeType() ast_pb.NodeType {
	return r.NodeType
}

// GetEntryId returns the entry contract ID.
func (r *RootSourceUnit) GetEntryId() int64 {
	return r.EntryContractId
}

// GetEntryName returns the entry contract name.
func (r *RootSourceUnit) GetEntryName() string {
	return r.EntryContractName
}

// GetContracts returns the list of contracts in the IR.
func (r *RootSourceUnit) GetContracts() []*Contract {
	return r.Contracts
}

// GetContractByName returns the contract with the given name from the IR.
// If no contract with the given name is found, it returns nil.
func (r *RootSourceUnit) GetContractByName(name string) *Contract {
	for _, su := range r.Contracts {
		if su.Name == name {
			return su
		}
	}

	return nil
}

// GetContractById returns the contract with the given ID from the IR.
// If no contract with the given ID is found, it returns nil.
func (r *RootSourceUnit) GetContractById(id int64) *Contract {
	for _, su := range r.Contracts {
		if su.Id == id {
			return su
		}
	}

	return nil
}

// GetEntryContract returns the entry contract from the IR.
func (r *RootSourceUnit) GetEntryContract() *Contract {
	return r.GetContractById(r.EntryContractId)
}

// HasContracts returns true if the AST has one or more contracts, false otherwise.
func (r *RootSourceUnit) HasContracts() bool {
	return len(r.Contracts) > 0
}

// GetContractsCount returns the count of contracts in the AST.
func (r *RootSourceUnit) GetContractsCount() int32 {
	return r.ContractsCount
}

// GetEips returns the EIPs discovered for any contract in the source units.
func (r *RootSourceUnit) GetEips() []*EIP {
	return r.Eips
}

// HasEips returns true if standard is already registered false otherwise.
func (r *RootSourceUnit) HasEIP(standard eip.Standard) bool {
	for _, e := range r.Eips {
		if e.Standard.Type == standard {
			return true
		}
	}

	return false
}

// GetContractTypes returns the list of contract types.
func (r *RootSourceUnit) GetContractTypes() []string {
	return r.ContractTypes
}

// HasContractType returns the list of contract types.
func (r *RootSourceUnit) HasContractType(ctype string) bool {
	for _, t := range r.ContractTypes {
		if t == ctype {
			return true
		}
	}

	return false
}

// SetContractType sets the contract type for the given standard.
func (r *RootSourceUnit) SetContractType(standard eip.Standard) {
	switch standard {
	case eip.EIP20:
		r.appendContractType("token")
	case eip.EIP721, eip.EIP1155:
		r.appendContractType("nft")
	case eip.EIP1967, eip.EIP1820:
		r.appendContractType("proxy")
		r.appendContractType("upgradeable")
	}
}

// appendContractType appends the given contract type to the list of contract types.
// It does not append if the contract type already exists in the list.
func (r *RootSourceUnit) appendContractType(contractType string) {
	if !r.HasContractType(contractType) {
		r.ContractTypes = append(r.ContractTypes, contractType)
	}
}

// ToProto is a placeholder function for converting the RootSourceUnit to a protobuf message.
func (r *RootSourceUnit) ToProto() *ir_pb.Root {
	proto := &ir_pb.Root{
		Id:                0,
		NodeType:          r.GetNodeType(),
		EntryContractId:   r.GetEntryId(),
		EntryContractName: r.GetEntryName(),
		ContractsCount:    r.GetContractsCount(),
		Contracts:         make([]*ir_pb.Contract, 0),
		ContractTypes:     r.GetContractTypes(),
	}

	for _, c := range r.GetContracts() {
		proto.Contracts = append(proto.Contracts, c.ToProto())
	}

	return proto
}

// processRoot processes the given root node of an AST and returns a RootSourceUnit.
// It populates the RootSourceUnit with the contracts from the AST.
func (b *Builder) processRoot(root *ast.RootNode) *RootSourceUnit {
	rootNode := &RootSourceUnit{
		unit:           root,
		NodeType:       root.GetType(),
		ContractsCount: int32(root.GetSourceUnitCount()),
		Contracts:      make([]*Contract, 0),
		ContractTypes:  make([]string, 0),
		Eips:           make([]*EIP, 0),
	}

	// No source units to process, so we're going to stop processing the root from here...
	if !root.HasSourceUnits() {
		return rootNode
	}

	entrySourceUnit := root.GetSourceUnitById(root.GetEntrySourceUnit())
	rootNode.EntryContractId = entrySourceUnit.GetId()
	rootNode.EntryContractName = entrySourceUnit.GetName()

	for _, su := range root.GetSourceUnits() {
		rootNode.Contracts = append(
			rootNode.Contracts,
			b.processContract(su),
		)
	}

	// Now is the time to process the EIPs
	b.processEips(rootNode)

	return rootNode
}