package ast

import (
	v3 "github.com/cncf/xds/go/xds/type/v3"
	ast_pb "github.com/txpull/protos/dist/go/ast"
	"github.com/txpull/solgo/parser"
)

type Contract struct {
	*ASTBuilder

	Id                      int64            `json:"id"`
	Name                    string           `json:"name"`
	NodeType                ast_pb.NodeType  `json:"node_type"`
	Src                     SrcNode          `json:"src"`
	Abstract                bool             `json:"abstract"`
	Kind                    ast_pb.NodeType  `json:"kind"`
	FullyImplemented        bool             `json:"fully_implemented"`
	Nodes                   []Node[NodeType] `json:"nodes"`
	LinearizedBaseContracts []int64          `json:"linearized_base_contracts"`
	BaseContracts           []*BaseContract  `json:"base_contracts"`
	ContractDependencies    []int64          `json:"contract_dependencies"`
}

func NewContractDefinition(b *ASTBuilder) *Contract {
	return &Contract{
		ASTBuilder: b,
	}
}

// SetReferenceDescriptor sets the reference descriptions of the Contract node.
func (c Contract) SetReferenceDescriptor(refId int64, refDesc *TypeDescription) bool {
	return false
}

func (c Contract) GetId() int64 {
	return c.Id
}

func (c Contract) GetType() ast_pb.NodeType {
	return c.NodeType
}

func (c Contract) GetSrc() SrcNode {
	return c.Src
}

func (c Contract) GetName() string {
	return c.Name
}

func (c Contract) IsAbstract() bool {
	return c.Abstract
}

func (c Contract) GetKind() ast_pb.NodeType {
	return c.Kind
}

func (c Contract) IsFullyImplemented() bool {
	return c.FullyImplemented
}

func (c Contract) GetNodes() []Node[NodeType] {
	return c.Nodes
}

func (c Contract) GetLinearizedBaseContracts() []int64 {
	return c.LinearizedBaseContracts
}

func (c Contract) GetBaseContracts() []*BaseContract {
	return c.BaseContracts
}

func (c Contract) GetContractDependencies() []int64 {
	return c.ContractDependencies
}

func (c Contract) GetTypeDescription() *TypeDescription {
	return nil
}

func (c Contract) ToProto() NodeType {
	proto := ast_pb.Contract{
		Id:                      c.Id,
		NodeType:                c.NodeType,
		Kind:                    c.Kind,
		Src:                     c.Src.ToProto(),
		Name:                    c.Name,
		Abstract:                c.Abstract,
		FullyImplemented:        c.FullyImplemented,
		LinearizedBaseContracts: c.LinearizedBaseContracts,
		ContractDependencies:    c.ContractDependencies,
		Nodes:                   make([]*v3.TypedStruct, 0),
		BaseContracts:           make([]*ast_pb.BaseContract, 0),
	}

	for _, baseContract := range c.BaseContracts {
		proto.BaseContracts = append(proto.BaseContracts, baseContract.ToProto())
	}

	for _, node := range c.Nodes {
		proto.Nodes = append(proto.Nodes, node.ToProto().(*v3.TypedStruct))
	}

	return NewTypedStruct(&proto, "Contract")
}

func (l Contract) Parse(unitCtx *parser.SourceUnitContext, ctx *parser.ContractDefinitionContext, rootNode *RootNode, unit *SourceUnit[Node[ast_pb.SourceUnit]]) {
	unit.Src = SrcNode{
		Id:          l.GetNextID(),
		Line:        int64(ctx.GetStart().GetLine()),
		Column:      int64(ctx.GetStart().GetColumn()),
		Start:       int64(ctx.GetStart().GetStart()),
		End:         int64(ctx.GetStop().GetStop()),
		Length:      int64(ctx.GetStop().GetStop() - ctx.GetStart().GetStart() + 1),
		ParentIndex: rootNode.Id,
	}

	// Set the absolute path of the source unit from provided sources map.
	// We are not dynamically loading files like the solc compiler does so we need to
	// provide the absolute path of the source unit from the sources map.
	unit.SetAbsolutePathFromSources(l.sources)
	unit.ExportedSymbols = append(unit.ExportedSymbols, Symbol{
		Id:           unit.Id,
		Name:         unit.Name,
		AbsolutePath: unit.AbsolutePath,
	})

	// Now we are going to resolve pragmas for current source unit...
	unit.Nodes = append(
		unit.Nodes,
		parsePragmasForSourceUnit(l.ASTBuilder, unitCtx, unit, nil, ctx, nil)...,
	)

	// Now we are going to resolve import paths for current source unit...
	nodeImports := parseImportPathsForSourceUnit(l.ASTBuilder, unitCtx, unit, nil, ctx, nil)
	unit.Nodes = append(
		unit.Nodes,
		nodeImports...,
	)

	contractNode := &Contract{
		Id:   l.GetNextID(),
		Name: ctx.Identifier().GetText(),
		Src: SrcNode{
			Line:        int64(ctx.GetStart().GetLine()),
			Column:      int64(ctx.GetStart().GetColumn()),
			Start:       int64(ctx.GetStart().GetStart()),
			End:         int64(ctx.GetStop().GetStop()),
			Length:      int64(ctx.GetStop().GetStop() - ctx.GetStart().GetStart() + 1),
			ParentIndex: unit.Id,
		},
		Abstract:                false,
		NodeType:                ast_pb.NodeType_CONTRACT_DEFINITION,
		Kind:                    ast_pb.NodeType_KIND_CONTRACT,
		LinearizedBaseContracts: make([]int64, 0),
		ContractDependencies:    make([]int64, 0),
		BaseContracts:           make([]*BaseContract, 0),
		Nodes:                   make([]Node[NodeType], 0),
		FullyImplemented:        true,
	}

	contractNode.BaseContracts = append(
		contractNode.BaseContracts,
		parseInheritanceFromCtx(
			l.ASTBuilder, unit, contractNode, ctx.InheritanceSpecifierList(),
		)...,
	)
	unit.BaseContracts = contractNode.BaseContracts

	contractNode.LinearizedBaseContracts = append(
		contractNode.LinearizedBaseContracts,
		contractNode.GetId(),
	)

	for _, nodeImport := range nodeImports {
		contractNode.LinearizedBaseContracts = append(
			contractNode.LinearizedBaseContracts,
			nodeImport.GetId(),
		)

		contractNode.ContractDependencies = append(
			contractNode.ContractDependencies,
			nodeImport.GetId(),
		)
	}

	for _, bodyElement := range ctx.AllContractBodyElement() {
		if bodyElement.IsEmpty() {
			contractNode.FullyImplemented = false
			continue
		}

		bodyNode := NewBodyNode(l.ASTBuilder)
		childNode := bodyNode.ParseDefinitions(unit, contractNode, bodyElement)
		if childNode != nil {
			contractNode.Nodes = append(
				contractNode.Nodes,
				childNode,
			)

			if bodyNode.NodeType == ast_pb.NodeType_FUNCTION_DEFINITION {
				if !bodyNode.Implemented {
					contractNode.FullyImplemented = false
				}
			}
		} else {
			contractNode.FullyImplemented = false
		}
	}

	unit.Nodes = append(unit.Nodes, contractNode)
	unit.Contract = contractNode
}