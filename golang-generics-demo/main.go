package main

import (
	"context"
	"fmt"
	"log"

	"github.com/rinkachi/golang-demos/golang-generics-demo/pkg/maps"
	"github.com/rinkachi/golang-demos/golang-generics-demo/pkg/ptr"
	"github.com/rinkachi/golang-demos/golang-generics-demo/pkg/repository"
	"github.com/rinkachi/golang-demos/golang-generics-demo/pkg/slices"
	"github.com/rinkachi/golang-demos/golang-generics-demo/pkg/structures"
)

// User Entity for Repository Demo
type User struct {
	ID   int
	Name string
}

func (u User) GetID() int { return u.ID }

func main() {
	fmt.Println("=== Generics Demo ===")

	// 1. Slices
	fmt.Println("\n--- Slices ---")
	nums := []int{1, 2, 3, 4, 5, 5, 2}
	result := slices.Map(slices.Unique(nums), func(n int) int { return n * 10 })
	fmt.Printf("Unique + Map (*10): %v\n", result)
	chunks := slices.Chunk(result, 2)
	fmt.Printf("Chunks (size 2): %v\n", chunks)

	// 2. Maps
	fmt.Println("\n--- Maps ---")
	data := map[string]int{"one": 1, "two": 2, "three": 3}
	keys := maps.Keys(data)
	fmt.Printf("Keys: %v\n", keys)

	// 3. Structures (Set)
	fmt.Println("\n--- Structures (Set) ---")
	set := structures.NewSet("a", "b", "a", "c")
	fmt.Printf("Set contains 'b': %v\n", set.Contains("b"))
	fmt.Printf("Set items: %v\n", set.Slice())

	// 4. Structures (Stack)
	fmt.Println("\n--- Structures (Stack) ---")
	stack := structures.NewStack[string]()
	stack.Push("first")
	stack.Push("second")
	val, _ := stack.Pop()
	fmt.Printf("Popped: %s\n", val)

	// 5. Ptr
	fmt.Println("\n--- Pointers ---")
	name := ptr.Of("Alice")
	fmt.Printf("Pointer value: %s\n", ptr.Unwrap(name))
	fmt.Printf("Nil default: %s\n", ptr.ValueOrDefault(nil, "Default"))

	// 6. Generic Repository
	fmt.Println("\n--- Generic Repository ---")
	repo := repository.NewInMemoryRepository[User, int]()
	ctx := context.Background()

	_ = repo.Create(ctx, User{ID: 1, Name: "Admin"})
	_ = repo.Create(ctx, User{ID: 2, Name: "User"})

	u, err := repo.FindByID(ctx, 1)
	if err == nil {
		fmt.Printf("Found User: %v\n", u)
	}

	all, _ := repo.FindAll(ctx)
	fmt.Printf("All Users Count: %d\n", len(all))
}
