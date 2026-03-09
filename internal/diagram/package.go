package diagram

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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

// GeneratePackageDiagram 生成包图
func GeneratePackageDiagram(rootDir, outputFile, title, format string) error {
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
		return fmt.Errorf("遍历目录失败：%w", err)
	}

	var output string
	switch format {
	case "mermaid":
		output = generateMermaidPackage(packages, moduleName, title)
	case "drawio":
		output = generateDrawIOPackage(packages, moduleName, title)
	default:
		output = generatePlantUMLPackage(packages, moduleName, title)
	}

	err = os.WriteFile(outputFile, []byte(output), 0644)
	if err != nil {
		return fmt.Errorf("写入文件失败：%w", err)
	}

	fmt.Printf("✅ 生成成功！\n")
	fmt.Printf("📄 输出文件：%s\n", outputFile)
	fmt.Printf("📊 共分析 %d 个包\n", len(packages))

	return nil
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

func generatePlantUMLPackage(packages map[string]*PackageNode, moduleName, title string) string {
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

func generateMermaidPackage(packages map[string]*PackageNode, moduleName, title string) string {
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

// PackageNodeDrawIO draw.io 包节点
type PackageNodeDrawIO struct {
	ID         string
	Name       string
	Structs    int
	Interfaces int
	Functions  int
	X          int
	Y          int
	Width      int
	Height     int
}

func generateDrawIOPackage(packages map[string]*PackageNode, moduleName, title string) string {
	var pkgNodes []PackageNodeDrawIO
	nodeID := 1

	var pkgPaths []string
	for path := range packages {
		pkgPaths = append(pkgPaths, path)
	}
	sort.Strings(pkgPaths)

	colWidth := 250
	rowHeight := 150
	columns := 4

	for i, path := range pkgPaths {
		pkg := packages[path]
		displayName := filepath.Base(path)
		if path == "." {
			displayName = "root"
		}

		col := i % columns
		row := i / columns

		pkgNodes = append(pkgNodes, PackageNodeDrawIO{
			ID:         strconv.Itoa(nodeID),
			Name:       displayName,
			Structs:    pkg.Structs,
			Interfaces: pkg.Interfaces,
			Functions:  pkg.Functions,
			X:          50 + col*colWidth,
			Y:          80 + row*rowHeight,
			Width:      200,
			Height:     100,
		})
		nodeID++
	}

	return generatePackageDrawIOXML(pkgNodes, packages, moduleName, title)
}

func generatePackageDrawIOXML(pkgNodes []PackageNodeDrawIO, packages map[string]*PackageNode, moduleName, title string) string {
	var sb strings.Builder

	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<mxfile host="app.diagrams.net" modified="2026-03-09T14:00:00.000Z" agent="uml-tools" etag="umltools" version="22.1.0" type="device">
  <diagram name="Package Diagram" id="package-diagram">
    <mxGraphModel dx="1422" dy="793" grid="1" gridSize="10" guides="1" tooltips="1" connect="1" arrows="1" fold="1" page="1" pageScale="1" pageWidth="1169" pageHeight="827" background="#ffffff" math="0" shadow="0">
      <root>
        <mxCell id="0"/>
        <mxCell id="1" parent="0"/>
`)

	sb.WriteString(fmt.Sprintf(`        <mxCell id="title" value="%s" style="text;html=1;strokeColor=none;fillColor=none;align=center;verticalAlign=middle;whiteSpace=wrap;rounded=0;fontSize=20;fontStyle=1" vertex="1" parent="1">
          <mxGeometry x="400" y="10" width="400" height="30" as="geometry"/>
        </mxCell>
`, title))

	for _, pkg := range pkgNodes {
		label := fmt.Sprintf("&lt;b&gt;%s&lt;/b&gt;&lt;hr/&gt;%d struct(s)&lt;br/&gt;%d interface(s)&lt;br/&gt;%d func(s)",
			pkg.Name, pkg.Structs, pkg.Interfaces, pkg.Functions)

		sb.WriteString(fmt.Sprintf(`        <mxCell id="pkg_%s" value="%s" style="shape=rectangle;rounded=1;html=1;whiteSpace=wrap;labelBackgroundColor=#ffffff;strokeColor=#000000;strokeWidth=1;fillColor=#ffffff;gradientColor=#ffffff;fontSize=12;align=left;spacingLeft=5;" vertex="1" parent="1">
          <mxGeometry x="%d" y="%d" width="%d" height="%d" as="geometry"/>
        </mxCell>
`, pkg.ID, label, pkg.X, pkg.Y, pkg.Width, pkg.Height))
	}

	sb.WriteString(`        <!-- 依赖关系 -->
`)

	edgeID := 1000
	seen := make(map[string]bool)
	for _, path := range getSortedKeys(packages) {
		pkg := packages[path]
		for _, imp := range pkg.Imports {
			relImp := strings.TrimPrefix(imp, moduleName+"/")
			if _, exists := packages[relImp]; exists {
				key := path + "->" + relImp
				if !seen[key] {
					seen[key] = true
					sourceID := getNodeID(pkgNodes, path)
					targetID := getNodeID(pkgNodes, relImp)
					if sourceID != "" && targetID != "" {
						sb.WriteString(fmt.Sprintf(`        <mxCell id="edge_%d" style="edgeStyle=orthogonalEdgeStyle;rounded=0;html=1;exitX=1;exitY=0.5;entryX=0;entryY=0.5;jettySize=auto;orthogonalLoop=1;" edge="1" parent="1" source="pkg_%s" target="pkg_%s">
          <mxGeometry relative="1" as="geometry"/>
        </mxCell>
`, edgeID, sourceID, targetID))
						edgeID++
					}
				}
			}
		}
	}

	sb.WriteString(`      </root>
    </mxGraphModel>
  </diagram>
</mxfile>
`)

	return sb.String()
}

func getNodeID(nodes []PackageNodeDrawIO, path string) string {
	name := filepath.Base(path)
	if path == "." {
		name = "root"
	}
	for _, node := range nodes {
		if node.Name == name {
			return node.ID
		}
	}
	return ""
}

func getSortedKeys(m map[string]*PackageNode) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
