package main

import (
	"fmt"
	"github.com/ldaidone/goembedx/pkg/embedx"
	"github.com/spf13/cobra"
)

var dbPath string

func Execute(engine *embedx.Embedder) {
	root := &cobra.Command{
		Use:   "goembedx",
		Short: "Vector embedding store and search engine",

		// attach engine to context for subcommands
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ctx := embedx.WithEngine(cmd.Context(), engine)
			cmd.SetContext(ctx)
		},
	}

	root.PersistentFlags().StringVar(&dbPath, "db", "./data", "database path")

	root.AddCommand(cmdInit(), cmdAdd(), cmdSearch())

	if err := root.Execute(); err != nil {
		panic(err)
	}
}

func cmdInit() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize vector store (already done in main)",
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

func cmdAdd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [id] [v1 v2 v3 ...]",
		Short: "Add vector",
		Args:  cobra.MinimumNArgs(2),
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

func cmdSearch() *cobra.Command {
	return &cobra.Command{
		Use:   "search [v1 v2 v3 ...]",
		Short: "Search vectors",
		Args:  cobra.MinimumNArgs(1),
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
