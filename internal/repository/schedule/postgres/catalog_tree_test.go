package postgres

import (
	"reflect"
	"testing"

	"github.com/GIT_USER_ID/GIT_REPO_ID/internal/domain"
)

func TestBuildCatalogRootPreservesNestedChildren(t *testing.T) {
	nodes := map[string]*domain.GroupCatalogNode{
		"group-b": {
			Abbr:       "IU5-12",
			UUID:       "group-b",
			NodeType:   "group",
			ParentUUID: "course",
			Course:     1,
		},
		"department": {
			Abbr:       "IU5",
			UUID:       "department",
			NodeType:   "department",
			ParentUUID: "faculty",
		},
		"root": {
			Abbr:     "Root",
			UUID:     "root",
			NodeType: "container",
		},
		"group-a": {
			Abbr:       "IU5-11",
			UUID:       "group-a",
			NodeType:   "group",
			ParentUUID: "course",
			Course:     1,
		},
		"course": {
			Abbr:       "IU5 (1 course)",
			UUID:       "course",
			NodeType:   "course",
			ParentUUID: "department",
			Course:     1,
		},
		"faculty": {
			Abbr:       "IU",
			UUID:       "faculty",
			NodeType:   "faculty",
			ParentUUID: "root",
		},
	}

	root := buildCatalogRoot(nodes)

	got := collectCatalogGroupUUIDsForTest(root)
	want := []string{"group-a", "group-b"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("group UUIDs = %#v, want %#v", got, want)
	}

	if len(root.Children) != 1 || root.Children[0].UUID != "faculty" {
		t.Fatalf("root children = %#v, want faculty child", root.Children)
	}
	faculty := root.Children[0]
	if len(faculty.Children) != 1 || faculty.Children[0].UUID != "department" {
		t.Fatalf("faculty children = %#v, want department child", faculty.Children)
	}
	department := faculty.Children[0]
	if len(department.Children) != 1 || department.Children[0].UUID != "course" {
		t.Fatalf("department children = %#v, want course child", department.Children)
	}
}

func TestBuildCatalogRootCombinesDisconnectedRoots(t *testing.T) {
	nodes := map[string]*domain.GroupCatalogNode{
		"root-b": {
			Abbr:     "B",
			UUID:     "root-b",
			NodeType: "faculty",
		},
		"group-b": {
			Abbr:       "B-11",
			UUID:       "group-b",
			NodeType:   "group",
			ParentUUID: "root-b",
		},
		"root-a": {
			Abbr:     "A",
			UUID:     "root-a",
			NodeType: "faculty",
		},
		"group-a": {
			Abbr:       "A-11",
			UUID:       "group-a",
			NodeType:   "group",
			ParentUUID: "root-a",
		},
	}

	root := buildCatalogRoot(nodes)

	if root.NodeType != "container" {
		t.Fatalf("root node type = %q, want container", root.NodeType)
	}

	gotRoots := []string{}
	for _, child := range root.Children {
		gotRoots = append(gotRoots, child.UUID)
	}
	wantRoots := []string{"root-a", "root-b"}
	if !reflect.DeepEqual(gotRoots, wantRoots) {
		t.Fatalf("root children UUIDs = %#v, want %#v", gotRoots, wantRoots)
	}

	gotGroups := collectCatalogGroupUUIDsForTest(root)
	wantGroups := []string{"group-a", "group-b"}
	if !reflect.DeepEqual(gotGroups, wantGroups) {
		t.Fatalf("group UUIDs = %#v, want %#v", gotGroups, wantGroups)
	}
}

func collectCatalogGroupUUIDsForTest(node domain.GroupCatalogNode) []string {
	result := []string{}
	if node.NodeType == "group" {
		result = append(result, node.UUID)
	}
	for _, child := range node.Children {
		result = append(result, collectCatalogGroupUUIDsForTest(child)...)
	}
	return result
}
