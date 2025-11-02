// Package main implements the goembedx command-line interface using Cobra.
// It provides commands for adding and searching vectors in the persistent store.
package main

import (
	"fmt"
	"github.com/ldaidone/goembedx/pkg/embedx"
	"github.com/spf13/cobra"
)

// dbPath stores the database path specified by the --db flag.
var dbPath string

// Execute runs the CLI command with the given embedx engine.
// It sets up the root command with subcommands and executes it.
func Execute(engine *embedx.Embedder) {
	root := &cobra.Command{
		Use:   "goembedx",
		Short: "Vector embedding store and search engine",
		Long: `goembedx is a lightweight local embedding store for Go.
It provides CLI tools for adding and searching vector embeddings.`,

		// attach engine to context for subcommands
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := embedx.WithEngine(cmd.Context(), engine)
			cmd.SetContext(ctx)
		},
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "./data", "database path for persistent storage")

	root.AddCommand(cmdInit(), cmdAdd(), cmdSearch())

	if err := root.Execute(); err != nil {
		panic(err)
	}
}

// cmdInit creates the 'init' command for initializing the vector store.
func cmdInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize vector store (already done in main)",
		Long:  `Initialize the vector store. This command confirms that the database path is ready for use.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := embedx.EngineFromContext(cmd.Context())
			if engine == nil {
				return fmt.Errorf("engine not available")
			}

			fmt.Println("Store ready at", dbPath)
			return nil
		},
	}
}

// cmdAdd creates the 'add' command for adding vectors to the store.
func cmdAdd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [id] [v1 v2 v3 ...]",
		Short: "Add vector",
		Long: `Add a vector with the given ID to the store.
The vector components should be provided as separate arguments after the ID.`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := embedx.FromContext(cmd.Context())
			if engine == nil {
				return fmt.Errorf("engine not initialized")
			}

			id := args[0]
			vec, err := parseFloat32Vec(args[1:])
			if err != nil {
				return err
			}

			if err := engine.Add(id, vec); err != nil {
				return err
			}

			fmt.Println("Vector added:", id)
			return nil
		},
	}
}

// cmdSearch creates the 'search' command for searching similar vectors.
func cmdSearch() *cobra.Command {
	return &cobra.Command{
		Use:   "search [v1 v2 v3 ...]",
		Short: "Search vectors",
		Long: `Search for vectors similar to the given query vector.
The query vector components should be provided as separate arguments.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			engine := embedx.FromContext(cmd.Context())
			if engine == nil {
				return fmt.Errorf("engine not initialized")
			}

			vec, err := parseFloat32Vec(args)
			if err != nil {
				return err
			}

			res, err := engine.Search(vec, 5) // default k=5 for now
			if err != nil {
				return err
			}

			fmt.Println("Results:")
			for _, r := range res {
				fmt.Printf("%s -> %.4f\n", r.ID, r.Score)
			}
			return nil
		},
	}
}

// parseFloat32Vec converts a slice of string representations to a slice of float32 values.
// It returns an error if any string cannot be parsed as a float32.
func parseFloat32Vec(strs []string) ([]float32, error) {
	vec := make([]float32, len(strs))
	for i, s := range strs {
		var f float64
		_, err := fmt.Sscan(s, &f)
		if err != nil {
			return nil, fmt.Errorf("invalid number: %s", s)
		}
		vec[i] = float32(f)
	}
	return vec, nil
}
