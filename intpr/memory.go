package intpr

import (
	"fmt"
	"simpl/errors"
	"simpl/tokens"
)

// Memory

type Memory struct {
	Size  int
	Ints  []map[string]int
	Bools []map[string]bool
	Funcs []map[string]*Function
}

func NewMemory() *Memory {
	return &Memory{Size: 1, Ints: []map[string]int{{}}, Bools: []map[string]bool{{}}, Funcs: []map[string]*Function{{}}}
}

func (m *Memory) Resize(scope int) {
    fmt.Println("Memory:")
    m.Print()
	for m.Size <= scope {
		m.Ints = append(m.Ints, map[string]int{})
		m.Bools = append(m.Bools, map[string]bool{})
		m.Funcs = append(m.Funcs, map[string]*Function{})
		m.Size++
	}
	if m.Size > scope+1 {
		m.Ints = m.Ints[:scope+1]
		m.Bools = m.Bools[:scope+1]
		m.Funcs = m.Funcs[:scope+1]
		m.Size = scope + 1
	}
}

func (m *Memory) GetBool(token tokens.Token) (bool, *errors.Error) {
	data := m.Bools
	for i := len(data) - 1; i >= 0; i-- {
		val, found := data[i][token.Value]
		if found {
			return val, nil
		}
	}
	return false, &errors.Error{Message: "the variable used to be here, but the memory got resized incorrectly", Type: errors.RuntimeError, Token: token}
}

func (m *Memory) GetInt(token tokens.Token) (int, *errors.Error) {
	data := m.Ints
	for i := len(data) - 1; i >= 0; i-- {
		val, found := data[i][token.Value]
		if found {
			return val, nil
		}
	}
	return 0, &errors.Error{Message: "the variable used to be here, but the memory got resized incorrectly", Type: errors.RuntimeError, Token: token}

}
func (m *Memory) GetFunc(token tokens.Token) (*Function, *errors.Error) {
	data := m.Funcs
	for i := len(data) - 1; i >= 0; i-- {
		val, found := data[i][token.Value]
		if found {
			return val, nil
		}
	}
	return nil, &errors.Error{Message: "the variable used to be here, but the memory got resized incorrectly", Type: errors.RuntimeError, Token: token}
}

func (m *Memory) SetInt(token tokens.Token, opToken tokens.Token, value int) {
	name := token.Value
	m.Ints[len(m.Ints)-1][name] = value
}

func (m *Memory) SetBool(token tokens.Token, opToken tokens.Token, value bool) {
	name := token.Value
	m.Bools[len(m.Bools)-1][name] = value
}

func (m *Memory) SetFunc(token tokens.Token, function *Function) {
	name := token.Value
	m.Funcs[len(m.Funcs)-1][name] = function
}

func (m *Memory) UpdateInt(token tokens.Token, value int) {
	name := token.Value
	for i := len(m.Ints) - 1; i >= 0; i-- {
		_, found := m.Ints[i][name]
		if found {
			m.Ints[i][name] = value
			break
		}
	}
}

func (m *Memory) IncInt(token tokens.Token, n int) {
	name := token.Value
	for i := len(m.Ints) - 1; i >= 0; i-- {
		_, found := m.Ints[i][name]
		if found {
			m.Ints[i][name] += n
			break
		}
	}
}

func (m *Memory) DecInt(token tokens.Token, n int) {
	name := token.Value
	for i := len(m.Ints) - 1; i >= 0; i-- {
		_, found := m.Ints[i][name]
		if found {
			m.Ints[i][name] -= n
			break
		}
	}
}

func (m *Memory) MulInt(token tokens.Token, n int) {
	name := token.Value
	for i := len(m.Ints) - 1; i >= 0; i-- {
		_, found := m.Ints[i][name]
		if found {
			m.Ints[i][name] *= n
			break
		}
	}
}

func (m *Memory) DivInt(token tokens.Token, n int) {
	name := token.Value
	for i := len(m.Ints) - 1; i >= 0; i-- {
		_, found := m.Ints[i][name]
		if found {
			m.Ints[i][name] /= n
			break
		}
	}
}

func (m *Memory) ModInt(token tokens.Token, n int) {
	name := token.Value
	for i := len(m.Ints) - 1; i >= 0; i-- {
		_, found := m.Ints[i][name]
		if found {
			m.Ints[i][name] %= n
			break
		}
	}
}

func (m *Memory) UpdateBool(token tokens.Token, value bool) {
	name := token.Value
	for i := len(m.Bools) - 1; i >= 0; i-- {
		_, found := m.Bools[i][name]
		if found {
			m.Bools[i][name] = value
			break
		}
	}
}

func (m *Memory) Print() {
	fmt.Println("Ints:")
	for _, data := range m.Ints {
		for k, v := range data {
			fmt.Printf("%s: %d\n", k, v)
		}
	}
	fmt.Println("Bools:")
	for _, data := range m.Bools {
		for k, v := range data {
			fmt.Printf("%s: %t\n", k, v)
		}
	}
    fmt.Println("Functions:")
    for _, data := range m.Funcs {
        for k, v := range data {
            fmt.Printf("%s ", k)
            v.Visualize()
        }
    }
}
