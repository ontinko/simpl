package memory

import (
	"fmt"
	"simpl/tokens"
)

// Memory

type Memory struct {
	ScopeCount int
	Ints       []map[string]int
	Bools      []map[string]bool
}

func NewMemory() *Memory {
	return &Memory{ScopeCount: 1, Ints: []map[string]int{{}}, Bools: []map[string]bool{{}}}
}

func (m *Memory) Resize(scope int) {
	for m.ScopeCount <= scope {
		m.Ints = append(m.Ints, map[string]int{})
		m.Bools = append(m.Bools, map[string]bool{})
		m.ScopeCount++
	}
	if m.ScopeCount > scope+1 {
		m.Ints = m.Ints[:scope+1]
		m.Bools = m.Bools[:scope+1]
		m.ScopeCount = scope
	}
}

func (m *Memory) GetBool(token tokens.Token) bool {
	data := m.Bools
	var result bool
	for i := len(data) - 1; i >= 0; i-- {
		val, found := data[i][token.Value]
		if found {
			result = val
		}
	}
	return result
}

func (m *Memory) GetInt(token tokens.Token) int {
	data := m.Ints
	var result int
	for i := len(data) - 1; i >= 0; i-- {
		val, found := data[i][token.Value]
		if found {
			result = val
			break
		}
	}
	return result
}

func (m *Memory) SetInt(token tokens.Token, opToken tokens.Token, value int) {
	name := token.Value
	m.Ints[len(m.Ints)-1][name] = value
}

func (m *Memory) SetBool(token tokens.Token, opToken tokens.Token, value bool) {
	name := token.Value
	m.Bools[len(m.Bools)-1][name] = value
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
}
