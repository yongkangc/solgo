package abis

import (
	"regexp"
	"strings"
)

// normalizeTypeName normalizes the type name in Solidity to its canonical form.
// For example, "uint" is normalized to "uint256", and "addresspayable" is normalized to "address".
// If the type name is not one of the special cases, it is returned as is.
func normalizeTypeName(typeName string) string {
	isArray, _ := regexp.MatchString(`\[\d+\]`, typeName)
	isSlice := strings.HasPrefix(typeName, "[]")

	switch {
	case isArray:
		numberPart := typeName[strings.Index(typeName, "[")+1 : strings.Index(typeName, "]")]
		typePart := typeName[strings.Index(typeName, "]")+1:]
		return "[" + numberPart + "]" + normalizeTypeName(typePart)

	case isSlice:
		typePart := typeName[2:]
		return "[]" + normalizeTypeName(typePart)

	case strings.HasPrefix(typeName, "uint"):
		if typeName == "uint" {
			return "uint256"
		}
		return typeName
	case strings.HasPrefix(typeName, "int"):
		if typeName == "int" {
			return "int256"
		}
		return typeName
	case strings.HasPrefix(typeName, "bool"):
		return typeName
	case strings.HasPrefix(typeName, "bytes"):
		return typeName
	case typeName == "string":
		return "string"
	case typeName == "address":
		return "address"
	case typeName == "addresspayable":
		return "address"
	case typeName == "tuple":
		return "tuple"
	default:
		return typeName
	}
}

// normalizeTypeNameWithStatus normalizes the type name in Solidity to its canonical form.
// For example, "uint" is normalized to "uint256", and "addresspayable" is normalized to "address".
// If the type name is not one of the special cases, it is returned as is with normalized == false for future manipulation...
func normalizeTypeNameWithStatus(typeName string) (string, bool) {
	isArray, _ := regexp.MatchString(`\[\d+\]`, typeName)
	isSlice := strings.HasPrefix(typeName, "[]")

	switch {
	case isArray:
		numberPart := typeName[strings.Index(typeName, "[")+1 : strings.Index(typeName, "]")]
		typePart := typeName[strings.Index(typeName, "]")+1:]

		return "[" + numberPart + "]" + normalizeTypeName(typePart), true

	case isSlice:
		typePart := typeName[2:]
		return "[]" + normalizeTypeName(typePart), true

	case strings.HasPrefix(typeName, "uint"):
		if typeName == "uint" {
			return "uint256", true
		}
		return typeName, true
	case strings.HasPrefix(typeName, "int"):
		if typeName == "int" {
			return "int256", true
		}
		return typeName, true
	case strings.HasPrefix(typeName, "bool"):
		return typeName, true
	case strings.HasPrefix(typeName, "bytes"):
		return typeName, true
	case typeName == "string":
		return "string", true
	case typeName == "address":
		return "address", true
	case typeName == "addresspayable":
		return "address", true
	case typeName == "tuple":
		return "tuple", true
	default:
		return typeName, false
	}
}

// normalizeStructTypeName normalizes the type name of a struct in Solidity to its canonical form.
// For example, "structName[]" is normalized to "tuple[]", and "structName" is normalized to "tuple".
func normalizeStructTypeName(definedStructs map[string]MethodIO, typeName string) string {
	switch {
	case strings.HasSuffix(typeName, "[]") && isStructType(definedStructs, strings.TrimSuffix(typeName, "[]")):
		// Handle array of structs
		return "tuple[]"
	default:
		return "tuple"
	}
}

// isMappingType checks if the given type name represents a mapping type in Solidity.
// It returns true if the type name contains the string "mapping", and false otherwise.
func isMappingType(name string) bool {
	return strings.Contains(name, "mapping")
}

// isStructType checks if a type name corresponds to a defined struct.
// definedStructs is a map from struct names to MethodIO objects representing the struct located in the AbiParser.
// Returns true if the type name corresponds to a defined struct, false otherwise.
func isStructType(definedStructs map[string]MethodIO, typeName string) bool {
	typeName = strings.TrimRight(typeName, "[]")
	_, exists := definedStructs[typeName]
	return exists
}

// isEnumType checks if a given type is an enumerated type.
// It takes a map of defined enums and a type name as arguments.
// The function returns true if the type name exists in the map of defined enums, indicating that it is an enumerated type.
// Otherwise, it returns false.
func isEnumType(definedEnums map[string]bool, typeName string) bool {
	_, exists := definedEnums[typeName]
	return exists
}

// isContractType checks if a given type is a contract type.
// It takes a map of defined contracts and a type name as arguments.
// The function returns true if the type name exists in the map of defined contracts, indicating that it is a contract type.
func isContractType(definedContracts map[string]ContractDefinition, typeName string) bool {
	_, exists := definedContracts[typeName]
	return exists
}

// isInterfaceType checks if a given type is an interface type.
// It takes a map of defined interfaces and a type name as arguments.
// The function returns true if the type name exists in the map of defined interfaces, indicating that it is an interface type.
func isInterfaceType(definedInterfaces map[string]bool, typeName string) bool {
	_, exists := definedInterfaces[typeName]
	return exists
}

// isLibraryType checks if a given type is a library type.
// It takes a map of defined libraries and a type name as arguments.
// The function returns true if the type name exists in the map of defined libraries, indicating that it is a library type.
func isLibraryType(definedLibraries map[string]bool, typeName string) bool {
	_, exists := definedLibraries[typeName]
	return exists
}

// parseMappingType parses a mapping type in Solidity ABI.
// It takes a string of the form "mapping(keyType => valueType)" and returns three values:
//   - A boolean indicating whether the parsing was successful. If the string is not a mapping type, this will be false.
//   - A slice of strings representing the types of the keys in the mapping. If the mapping is nested, this will contain multiple elements.
//   - A slice of strings representing the types of the values in the mapping. If the mapping is nested, the inner mappings will be flattened,
//     and this will contain the types of the innermost values.
func parseMappingType(abi string) (bool, []string, []string) {
	re := regexp.MustCompile(`mapping\((\w+)\s*=>\s*(.+)\)`)
	matches := re.FindStringSubmatch(abi)

	if len(matches) < 3 {
		return false, []string{}, []string{}
	}

	input := []string{matches[1]}
	output := []string{matches[2]}

	// If the output is another mapping, parse it recursively
	if isMappingType(output[0]) {
		_, nestedInput, nestedOutput := parseMappingType(output[0])
		input = append(input, nestedInput...)
		output = nestedOutput
	}

	return true, input, output
}