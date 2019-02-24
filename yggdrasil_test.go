package raizel

import (
	"fmt"
	"testing"

	"github.com/rjansen/yggdrasil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testRegister struct {
	name       string
	repository Repository
	err        error
}

func TestRegister(test *testing.T) {
	scenarios := []testRegister{
		{
			name:       "Register the Repository reference",
			repository: repositoryMock{},
		},
		{
			name:       "Register a nil Repository reference",
			repository: nil,
		},
	}

	for index, scenario := range scenarios {
		test.Run(
			fmt.Sprintf("[%d]-%s", index, scenario.name),
			func(t *testing.T) {
				roots := yggdrasil.NewRoots()
				err := Register(&roots, scenario.repository)
				assert.Equal(t, scenario.err, err)

				tree := roots.NewTreeDefault()
				repository, err := tree.Reference(repositoryPath)

				require.Nil(t, err, "tree reference error")
				require.Exactly(t, scenario.repository, repository, "repository reference")
			},
		)
	}
}

type testReference struct {
	name       string
	references map[yggdrasil.Path]yggdrasil.Reference
	tree       yggdrasil.Tree
	err        error
}

func (scenario *testReference) setup(t *testing.T) {
	roots := yggdrasil.NewRoots()
	for path, reference := range scenario.references {
		err := roots.Register(path, reference)
		assert.Nil(t, err, "register error")
	}
	scenario.tree = roots.NewTreeDefault()
}

func TestReference(test *testing.T) {
	scenarios := []testReference{
		{
			name: "Access the Repository Reference",
			references: map[yggdrasil.Path]yggdrasil.Reference{
				repositoryPath: yggdrasil.NewReference(repositoryMock{}),
			},
		},
		{
			name: "Access a nil Repository Reference",
			references: map[yggdrasil.Path]yggdrasil.Reference{
				repositoryPath: yggdrasil.NewReference(nil),
			},
		},
		{
			name: "When Repository was not register returns path not found",
			err:  yggdrasil.ErrPathNotFound,
		},
		{
			name: "When a invalid Repository was register returns invalid reference error",
			references: map[yggdrasil.Path]yggdrasil.Reference{
				repositoryPath: yggdrasil.NewReference(new(struct{})),
			},
			err: ErrInvalidReference,
		},
	}

	for index, scenario := range scenarios {
		test.Run(
			fmt.Sprintf("[%d]-%s", index, scenario.name),
			func(t *testing.T) {
				scenario.setup(t)

				_, err := Reference(scenario.tree)
				assert.Equal(t, scenario.err, err, "reference error")
				if scenario.err != nil {
					assert.PanicsWithValue(t, scenario.err,
						func() {
							_ = MustReference(scenario.tree)
						},
					)
				} else {
					assert.NotPanics(t,
						func() {
							_ = MustReference(scenario.tree)
						},
					)
				}
			},
		)
	}
}