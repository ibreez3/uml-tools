package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PackageInfo 存储包的信息
type PackageInfo struct {
	Name       string
	FilePath   string
	Structs    []StructInfo
	Interfaces []InterfaceInfo
	Imports    []string
}

// StructInfo 存储结构体信息
type StructInfo struct {
	Name    string
	Fields  []FieldInfo
	Methods []MethodInfo
}

// InterfaceInfo 存储接口信息
type InterfaceInfo struct {
	Name    string
	Methods []MethodInfo
}

// FieldInfo 存储字段信息
type FieldInfo struct {
	Name string
	Type string
}

// MethodInfo 存储方法信息
type MethodInfo struct {
	Name   string
	Params string
	Return string
}

func main() {
	outputFile := flag.String("o", "classDiagram.puml", "输出文件路径")
	title := flag.String("title", "Go Project Class Diagram", "图表标题")
	format := flag.String("format", "plantuml", "输出格式：plantuml 或 mermaid")
	flag.Parse()

	rootDir := "."
	if len(flag.Args()) > 0 {
		rootDir = flag.Args()[0]
	}

	fmt.Printf("🔍 开始分析项目：%s\n", filepath.Abs(rootDir))

	packages := make(map[string]*PackageInfo)

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, "_test.go") ||
			strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/.git/") ||
			strings.Contains(path, "/node_modules/") {
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Printf("⚠️  解析失败 %s: %v\n", path, err)
			return nil
		}

		pkgName := file.Name.Name
		pkgPath := filepath.Dir(path)

		relPath, _ := filepath.Rel(rootDir, pkgPath)
		if relPath == "." {
			relPath = pkgName
		} else {
			relPath = filepath.Join(relPath, pkgName)
		}

		pkg, exists := packages[relPath]
		if !exists {
			pkg = &PackageInfo{
				Name:       pkgName,
				FilePath:   relPath,
				Structs:    []StructInfo{},
				Interfaces: []InterfaceInfo{},
				Imports:    []string{},
			}
			packages[relPath] = pkg
		}

		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			pkg.Imports = append(pkg.Imports, importPath)
		}

		// 提取结构体和接口
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}

			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					structInfo := StructInfo{
						Name:   typeSpec.Name.Name,
						Fields: []FieldInfo{},
					}

					if structType.Fields != nil {
						for _, field := range structType.Fields.List {
							fieldType := formatType(field.Type)
							if len(field.Names) > 0 {
								for _, name := range field.Names {
									structInfo.Fields = append(structInfo.Fields, FieldInfo{
										Name: name.Name,
										Type: fieldType,
									})
								}
							} else {
								structInfo.Fields = append(structInfo.Fields, FieldInfo{
									Name: "",
									Type: fieldType,
								})
							}
						}
					}

					pkg.Structs = append(pkg.Structs, structInfo)
				}

				if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					interfaceInfo := InterfaceInfo{
						Name:    typeSpec.Name.Name,
						Methods: []MethodInfo{},
					}

					if interfaceType.Methods != nil {
						for _, method := range interfaceType.Methods.List {
							if len(method.Names) > 0 {
								methodName := method.Names[0].Name
								params, returns := formatFuncType(method.Type)
								interfaceInfo.Methods = append(interfaceInfo.Methods, MethodInfo{
									Name:   methodName,
									Params: params,
									Return: returns,
								})
							}
						}
					}

					pkg.Interfaces = append(pkg.Interfaces, interfaceInfo)
				}
			}
		}

		// 提取方法
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcDecl.Recv != nil && len(funcDecl.Recv.List) > 0 {
				recv := funcDecl.Recv.List[0]
				var receiverName string

				switch t := recv.Type.(type) {
				case *ast.Ident:
					receiverName = t.Name
				case *ast.StarExpr:
					if ident, ok := t.X.(*ast.Ident); ok {
						receiverName = ident.Name
					}
				}

				if receiverName != "" {
					params, returns := formatFuncType(funcDecl.Type)

					for i := range pkg.Structs {
						if pkg.Structs[i].Name == receiverName {
							pkg.Structs[i].Methods = append(pkg.Structs[i].Methods, MethodInfo{
								Name:   funcDecl.Name.Name,
								Params: params,
								Return: returns,
							})
							break
						}
					}

					for i := range pkg.Interfaces {
						if pkg.Interfaces[i].Name == receiverName {
							pkg.Interfaces[i].Methods = append(pkg.Interfaces[i].Methods, MethodInfo{
								Name:   funcDecl.Name.Name,
								Params: params,
								Return: returns,
							})
							break
						}
					}
				}
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("❌ 遍历目录失败：%v\n", err)
		os.Exit(1)
	}

	var output string
	if *format == "mermaid" {
		output = generateMermaid(packages, *title)
	} else {
		output = generatePlantUML(packages, *title)
	}

	err = os.WriteFile(*outputFile, []byte(output), 0644)
	if err != nil {
		fmt.Printf("❌ 写入文件失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 生成成功！\n")
	fmt.Printf("📄 输出文件：%s\n", *outputFile)
	fmt.Printf("📊 共分析 %d 个包\n", len(packages))

	totalStructs := 0
	totalInterfaces := 0
	for _, pkg := range packages {
		totalStructs += len(pkg.Structs)
		totalInterfaces += len(pkg.Interfaces)
	}
	fmt.Printf("📦 发现 %d 个结构体，%d 个接口\n", totalStructs, totalInterfaces)
}

func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatType(t.Key), formatType(t.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", formatType(t.X), t.Sel.Name)
	case *ast.FuncType:
		return "func(...)"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.ChanType:
		return "chan " + formatType(t.Value)
	default:
		return "unknown"
	}
}

func formatFuncType(funcType *ast.FuncType) (params string, returns string) {
	var paramList []string
	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			paramType := formatType(param.Type)
			if len(param.Names) > 0 {
				for _, name := range param.Names {
					paramList = append(paramList, fmt.Sprintf("%s %s", name.Name, paramType))
				}
			} else {
				paramList = append(paramList, paramType)
			}
		}
	}
	params = strings.Join(paramList, ", ")

	var returnList []string
	if funcType.Results != nil {
		for _, result := range funcType.Results.List {
			returnType := formatType(result.Type)
			if len(result.Names) > 0 {
				for _, name := range result.Names {
					returnList = append(returnList, fmt.Sprintf("%s %s", name.Name, returnType))
				}
			} else {
				returnList = append(returnList, returnType)
			}
		}
	}
	returns = strings.Join(returnList, ", ")

	return
}

func generatePlantUML(packages map[string]*PackageInfo, title string) string {
	var sb strings.Builder

	sb.WriteString("@startuml\n")
	sb.WriteString("title " + title + "\n")
	sb.WriteString("skinparam namespaceSeparator ::\n\n")

	var pkgNames []string
	for pkgName := range packages {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	for _, pkgName := range pkgNames {
		pkg := packages[pkgName]

		if len(pkg.Structs) == 0 && len(pkg.Interfaces) == 0 {
			continue
		}

		namespace := strings.ReplaceAll(pkgName, "/", ".")
		sb.WriteString(fmt.Sprintf("namespace %s {\n", namespace))

		for _, s := range pkg.Structs {
			sb.WriteString(fmt.Sprintf("    class %s {\n", s.Name))
			for _, f := range s.Fields {
				visibility := "+"
				if f.Name != "" && !strings.HasPrefix(f.Name, strings.ToUpper(string(f.Name[0]))) {
					visibility = "-"
				}
				if f.Name == "" {
					sb.WriteString(fmt.Sprintf("        %s %s\n", visibility, f.Type))
				} else {
					sb.WriteString(fmt.Sprintf("        %s %s : %s\n", visibility, f.Name, f.Type))
				}
			}
			for _, m := range s.Methods {
				sb.WriteString(fmt.Sprintf("        + %s(%s) %s\n", m.Name, m.Params, m.Return))
			}
			sb.WriteString("    }\n\n")
		}

		for _, i := range pkg.Interfaces {
			sb.WriteString(fmt.Sprintf("    interface %s {\n", i.Name))
			for _, m := range i.Methods {
				sb.WriteString(fmt.Sprintf("        + %s(%s) %s\n", m.Name, m.Params, m.Return))
			}
			sb.WriteString("    }\n\n")
		}

		sb.WriteString("}\n\n")
	}

	sb.WriteString("' 关系（可以根据需要手动补充）\n")
	for _, pkgName := range pkgNames {
		pkg := packages[pkgName]
		namespace := strings.ReplaceAll(pkgName, "/", ".")
		for _, s := range pkg.Structs {
			for _, f := range s.Fields {
				if !strings.Contains(f.Type, ".") && f.Type != "string" && f.Type != "int" &&
					f.Type != "int64" && f.Type != "bool" && f.Type != "error" &&
					f.Type != "[]byte" && f.Type != "map" && f.Type != "func" {
					sb.WriteString(fmt.Sprintf("%s.%s --> %s\n", namespace, s.Name, f.Type))
				}
			}
		}
	}

	sb.WriteString("@enduml\n")
	return sb.String()
}

func generateMermaid(packages map[string]*PackageInfo, title string) string {
	var sb strings.Builder

	sb.WriteString("%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#ffdfb3', 'edgeLabelBackground':'#fff'}}}%%\n")
	sb.WriteString("classDiagram\n")
	sb.WriteString("    title " + title + "\n\n")

	var pkgNames []string
	for pkgName := range packages {
		pkgNames = append(pkgNames, pkgName)
	}
	sort.Strings(pkgNames)

	for _, pkgName := range pkgNames {
		pkg := packages[pkgName]
		namespace := strings.ReplaceAll(pkgName, "/", ".")

		for _, s := range pkg.Structs {
			sb.WriteString(fmt.Sprintf("    class %s:::%s {\n", namespace+"."+s.Name, sanitizeName(pkgName)))
			for _, f := range s.Fields {
				if f.Name != "" {
					sb.WriteString(fmt.Sprintf("        %s %s\n", f.Name, f.Type))
				}
			}
			for _, m := range s.Methods {
				sb.WriteString(fmt.Sprintf("        %s(%s) %s\n", m.Name, m.Params, m.Return))
			}
			sb.WriteString("    }\n\n")
		}

		for _, i := range pkg.Interfaces {
			sb.WriteString(fmt.Sprintf("    class %s:::%s {\n", namespace+"."+i.Name, sanitizeName(pkgName)))
			for _, m := range i.Methods {
				sb.WriteString(fmt.Sprintf("        <<interface>>\n"))
				sb.WriteString(fmt.Sprintf("        %s(%s) %s\n", m.Name, m.Params, m.Return))
			}
			sb.WriteString("    }\n\n")
		}
	}

	sb.WriteString("    %% 样式定义\n")
	for i, pkgName := range pkgNames {
		color := []string{"#ffdfb3", "#b3d9ff", "#b3ffb3", "#ffb3b3", "#d9b3ff"}[i%5]
		sb.WriteString(fmt.Sprintf("    classDef %s fill:%s\n", sanitizeName(pkgName), color))
	}

	return sb.String()
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	return name
}
