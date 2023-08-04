package ast

import (
	"fmt"

	ast_pb "github.com/txpull/protos/dist/go/ast"
	"github.com/txpull/solgo/parser"
)

type EnumDefinition struct {
	*ASTBuilder
	SourceUnitName  string           `json:"-"`
	Id              int64            `json:"id"`
	NodeType        ast_pb.NodeType  `json:"node_type"`
	Src             SrcNode          `json:"src"`
	Name            string           `json:"name"`
	CanonicalName   string           `json:"canonical_name"`
	TypeDescription *TypeDescription `json:"type_description"`
	Members         []Node[NodeType] `json:"members"`
}

func NewEnumDefinition(b *ASTBuilder) *EnumDefinition {
	return &EnumDefinition{
		ASTBuilder: b,
		Id:         b.GetNextID(),
		NodeType:   ast_pb.NodeType_ENUM_DEFINITION,
	}
}

// SetReferenceDescriptor sets the reference descriptions of the EnumDefinition node.
// We don't need to do any reference description updates here, at least for now...
func (e *EnumDefinition) SetReferenceDescriptor(refId int64, refDesc *TypeDescription) bool {
	return false
}

func (e *EnumDefinition) GetId() int64 {
	return e.Id
}

func (e *EnumDefinition) GetType() ast_pb.NodeType {
	return e.NodeType
}

func (e *EnumDefinition) GetSrc() SrcNode {
	return e.Src
}

func (e *EnumDefinition) GetName() string {
	return e.Name
}

func (e *EnumDefinition) GetTypeDescription() *TypeDescription {
	return e.TypeDescription
}

func (e *EnumDefinition) GetCanonicalName() string {
	return e.CanonicalName
}

func (e *EnumDefinition) GetMembers() []*Parameter {
	toReturn := make([]*Parameter, 0)

	for _, member := range e.Members {
		toReturn = append(toReturn, member.(*Parameter))
	}

	return toReturn
}

func (e *EnumDefinition) GetSourceUnitName() string {
	return e.SourceUnitName
}

func (e *EnumDefinition) ToProto() NodeType {
	proto := ast_pb.Enum{
		Id:              e.GetId(),
		Name:            e.GetName(),
		CanonicalName:   e.GetCanonicalName(),
		NodeType:        e.GetType(),
		Src:             e.GetSrc().ToProto(),
		Members:         make([]*ast_pb.Parameter, 0),
		TypeDescription: e.GetTypeDescription().ToProto(),
	}

	for _, member := range e.GetMembers() {
		proto.Members = append(
			proto.Members,
			member.ToProto().(*ast_pb.Parameter),
		)
	}

	return NewTypedStruct(&proto, "Enum")
}

func (e *EnumDefinition) GetNodes() []Node[NodeType] {
	return e.Members
}

func (e *EnumDefinition) Parse(
	unit *SourceUnit[Node[ast_pb.SourceUnit]],
	contractNode Node[NodeType],
	bodyCtx parser.IContractBodyElementContext,
	ctx *parser.EnumDefinitionContext,
) Node[NodeType] {
	e.Src = SrcNode{
		Id:          e.GetNextID(),
		Line:        int64(ctx.GetStart().GetLine()),
		Column:      int64(ctx.GetStart().GetColumn()),
		Start:       int64(ctx.GetStart().GetStart()),
		End:         int64(ctx.GetStop().GetStop()),
		Length:      int64(ctx.GetStop().GetStop() - ctx.GetStart().GetStart()),
		ParentIndex: contractNode.GetId(),
	}
	e.SourceUnitName = unit.GetName()
	e.Name = ctx.GetName().GetText()
	e.CanonicalName = fmt.Sprintf("%s.%s", unit.GetName(), e.Name)
	e.TypeDescription = &TypeDescription{
		TypeIdentifier: fmt.Sprintf("t_enum_$_%s_$%d", e.Name, e.Id),
		TypeString:     fmt.Sprintf("enum %s", e.CanonicalName),
	}

	for _, enumCtx := range ctx.GetEnumValues() {
		id := e.GetNextID()
		e.Members = append(
			e.Members,
			&Parameter{
				Id: id,
				Src: SrcNode{
					Line:        int64(enumCtx.GetStart().GetLine()),
					Column:      int64(enumCtx.GetStart().GetColumn()),
					Start:       int64(enumCtx.GetStart().GetStart()),
					End:         int64(enumCtx.GetStop().GetStop()),
					Length:      int64(enumCtx.GetStop().GetStop() - enumCtx.GetStart().GetStart()),
					ParentIndex: e.Id,
				},
				Name:     enumCtx.GetText(),
				NodeType: ast_pb.NodeType_ENUM_VALUE,
				TypeDescription: &TypeDescription{
					TypeIdentifier: fmt.Sprintf("t_enum_$_%s$_%s_$%d", e.Name, enumCtx.GetText(), id),
					TypeString:     fmt.Sprintf("enum %s.%s", e.CanonicalName, enumCtx.GetText()),
				},
			},
		)
	}
	e.currentEnums = append(e.currentEnums, e)

	return e
}