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

// PackageNode 表示一个包节点
type PackageNode struct {
	Name       string
	FilePath   string
	Imports    []string
	Structs    int
	Interfaces int
	Functions  int
}

func main() {
	outputFile := flag.String("o", "packageDiagram.puml", "输出文件路径")
	title := flag.String("title", "Go Project Package Diagram", "图表标题")
	format := flag.String("format", "plantuml", "输出格式：plantuml 或 mermaid")
	flag.Parse()

	rootDir := "."
	if len(flag.Args()) > 0 {
		rootDir = flag.Args()[0]
	}

	fmt.Printf("🔍 开始分析项目：%s\n", filepath.Abs(rootDir))

	packages := make(map[string]*PackageNode)

	moduleName := getModuleName(rootDir)
	if moduleName != "" {
		fmt.Printf("📦 模块名：%s\n", moduleName)
	}

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
			return nil
		}

		pkgName := file.Name.Name
		pkgPath := filepath.Dir(path)

		relPath, _ := filepath.Rel(rootDir, pkgPath)
		if relPath == "." {
			relPath = pkgName
		}

		pkg, exists := packages[relPath]
		if !exists {
			pkg = &PackageNode{
				Name:       pkgName,
				FilePath:   relPath,
				Imports:    []string{},
				Structs:    0,
				Interfaces: 0,
				Functions:  0,
			}
			packages[relPath] = pkg
		}

		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if moduleName != "" && strings.HasPrefix(importPath, moduleName) {
				pkg.Imports = append(pkg.Imports, importPath)
			}
		}

		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						switch ts.Type.(type) {
						case *ast.StructType:
							pkg.Structs++
						case *ast.InterfaceType:
							pkg.Interfaces++
						}
					}
				}
			case *ast.FuncDecl:
				if d.Recv == nil {
					pkg.Functions++
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
		output = generateMermaid(packages, moduleName, *title)
	} else {
		output = generatePlantUML(packages, moduleName, *title)
	}

	err = os.WriteFile(*outputFile, []byte(output), 0644)
	if err != nil {
		fmt.Printf("❌ 写入文件失败：%v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ 生成成功！\n")
	fmt.Printf("📄 输出文件：%s\n", *outputFile)
	fmt.Printf("📊 共分析 %d 个包\n", len(packages))
}

func getModuleName(rootDir string) string {
	goModPath := filepath.Join(rootDir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func generatePlantUML(packages map[string]*PackageNode, moduleName, title string) string {
	var sb strings.Builder

	sb.WriteString("@startuml\n")
	sb.WriteString("title " + title + "\n")
	sb.WriteString("skinparam packageStyle rectangle\n")
	sb.WriteString("skinparam backgroundColor #EEEBDC\n")
	sb.WriteString("skinparam packageBorderColor #333333\n")
	sb.WriteString("skinparam packageBackgroundColor #FFFFFF\n\n")

	var pkgPaths []string
	for path := range packages {
		pkgPaths = append(pkgPaths, path)
	}
	sort.Strings(pkgPaths)

	sb.WriteString("' 包定义\n")
	for _, path := range pkgPaths {
		pkg := packages[path]
		displayName := filepath.Base(path)
		if path == "." {
			displayName = "root"
		}

		sb.WriteString(fmt.Sprintf("package \"%s\" as %s {\n", displayName, sanitizeName(path)))
		sb.WriteString(fmt.Sprintf("  [%d struct(s)\\n%d interface(s)\\n%d func(s)]\n",
			pkg.Structs, pkg.Interfaces, pkg.Functions))
		sb.WriteString("}\n\n")
	}

	sb.WriteString("' 依赖关系\n")
	seen := make(map[string]bool)
	for _, path := range pkgPaths {
		pkg := packages[path]
		for _, imp := range pkg.Imports {
			relImp := strings.TrimPrefix(imp, moduleName+"/")
			if _, exists := packages[relImp]; exists {
				key := path + "->" + relImp
				if !seen[key] {
					seen[key] = true
					sb.WriteString(fmt.Sprintf("%s ..> %s : imports\n",
						sanitizeName(path), sanitizeName(relImp)))
				}
			}
		}
	}

	sb.WriteString("@enduml\n")
	return sb.String()
}

func generateMermaid(packages map[string]*PackageNode, moduleName, title string) string {
	var sb strings.Builder

	sb.WriteString("%%{init: {'theme': 'base', 'themeVariables': { 'primaryColor': '#ffdfb3', 'edgeLabelBackground':'#fff', 'tertiaryColor': '#fff5e6'}}}%%\n")
	sb.WriteString("graph TD\n")
	sb.WriteString("    title[" + title + "]\n\n")

	var pkgPaths []string
	for path := range packages {
		pkgPaths = append(pkgPaths, path)
	}
	sort.Strings(pkgPaths)

	for _, path := range pkgPaths {
		pkg := packages[path]
		displayName := filepath.Base(path)
		if path == "." {
			displayName = "root"
		}

		nodeId := sanitizeName(path)
		label := fmt.Sprintf("%s\\n[%dS/%dI/%dF]", displayName, pkg.Structs, pkg.Interfaces, pkg.Functions)
		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]\n", nodeId, label))
	}

	seen := make(map[string]bool)
	for _, path := range pkgPaths {
		pkg := packages[path]
		for _, imp := range pkg.Imports {
			relImp := strings.TrimPrefix(imp, moduleName+"/")
			if _, exists := packages[relImp]; exists {
				key := path + "->" + relImp
				if !seen[key] {
					seen[key] = true
					sb.WriteString(fmt.Sprintf("    %s --> %s\n",
						sanitizeName(path), sanitizeName(relImp)))
				}
			}
		}
	}

	return sb.String()
}

func sanitizeName(name string) string {
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	if name == "." {
		return "root_pkg"
	}
	return name
}
