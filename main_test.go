package main

import (
	"go/ast"
	"go/types"
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestNewFunctionInfo(t *testing.T) {
	t.Run("TestNewFunctionInfo_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var funcDecl *ast.FuncDecl
		var pkg *packages.Package

		result0 := NewFunctionInfo(funcDecl, pkg)

		var expect0 *FuncInfo
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestRequiredImports(t *testing.T) {
	t.Run("TestRequiredImports_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.RequiredImports()

		var expect0 []string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintDefaultArgs(t *testing.T) {
	t.Run("TestPrintDefaultArgs_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintDefaultArgs()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintDefaultReturns(t *testing.T) {
	t.Run("TestPrintDefaultReturns_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintDefaultReturns()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintReceiverCtor(t *testing.T) {
	t.Run("TestPrintReceiverCtor_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintReceiverCtor()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintCall(t *testing.T) {
	t.Run("TestPrintCall_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintCall()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintDefaultExpects(t *testing.T) {
	t.Run("TestPrintDefaultExpects_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintDefaultExpects()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintDefaultVarArgs(t *testing.T) {
	t.Run("TestPrintDefaultVarArgs_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintDefaultVarArgs()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintArgsAsStructFields(t *testing.T) {
	t.Run("TestPrintArgsAsStructFields_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintArgsAsStructFields()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintTableArgs(t *testing.T) {
	t.Run("TestPrintTableArgs_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver FuncInfo

		result0 := receiver.PrintTableArgs()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestMain(t *testing.T) {
	t.Run("TestMain_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		main()

	})
}

func TestDefaultFail(t *testing.T) {
	t.Run("TestDefaultFail_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver TestTemplateData

		result0 := receiver.DefaultFail()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestPrintDefaultCtor(t *testing.T) {
	t.Run("TestPrintDefaultCtor_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver Arg

		result0 := receiver.PrintDefaultCtor()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestExtractPackagePrefix(t *testing.T) {
	t.Run("TestExtractPackagePrefix_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var typeName string

		result0, result1, result2 := extractPackagePrefix(typeName)

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

		var expect1 string
		if !reflect.DeepEqual(result1, expect1) {
			t.Errorf("Expected %v, got %v", expect1, result1)
		}

		var expect2 string
		if !reflect.DeepEqual(result2, expect2) {
			t.Errorf("Expected %v, got %v", expect2, result2)
		}

	})
}

func TestIsStandardLibrary(t *testing.T) {
	t.Run("TestIsStandardLibrary_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var pkg *packages.Package
		var goroot string

		result0 := isStandardLibrary(pkg, goroot)

		var expect0 bool
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestRun(t *testing.T) {
	t.Run("TestRun_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var asTable bool

		run(asTable)

	})
}

func TestGoFmt(t *testing.T) {
	t.Run("TestGoFmt_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var path string

		result0 := goFmt(path)

		var expect0 error
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestGetSimplifiedTypeName(t *testing.T) {
	t.Run("TestGetSimplifiedTypeName_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var typ types.Type
		var targetTypesPkg *types.Package

		result0 := getSimplifiedTypeName(typ, targetTypesPkg)

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestGetQualifiedTypeName(t *testing.T) {
	t.Run("TestGetQualifiedTypeName_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var typ types.Type

		result0 := getQualifiedTypeName(typ)

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestNewPackageInfo(t *testing.T) {
	t.Run("TestNewPackageInfo_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var pkgPath string

		result0 := NewPackageInfo(pkgPath)

		var expect0 PackageInfo
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestTestableFuncs(t *testing.T) {
	t.Run("TestTestableFuncs_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver PackageInfo

		result0 := receiver.TestableFuncs()

		var expect0 []FuncInfo
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestAdd(t *testing.T) {
	t.Run("TestAdd_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver *PackageInfo
		var fi *FuncInfo

		receiver.Add(fi)

	})
}

func TestPrintImports(t *testing.T) {
	t.Run("TestPrintImports_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		var receiver PackageInfo

		result0 := receiver.PrintImports()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

	})
}

func TestGetGOROOT(t *testing.T) {
	t.Run("TestGetGOROOT_0", func(t *testing.T) {

		// delete this after your implementation
		t.Fatalf("test not implemented")

		result0, result1 := getGOROOT()

		var expect0 string
		if !reflect.DeepEqual(result0, expect0) {
			t.Errorf("Expected %v, got %v", expect0, result0)
		}

		var expect1 error
		if !reflect.DeepEqual(result1, expect1) {
			t.Errorf("Expected %v, got %v", expect1, result1)
		}

	})
}
