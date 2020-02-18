/*
Nanostate is loaded by Id or filename.
*/

package nanostate

import (
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type NanoStateMeta struct {
	Id       string
	Filename string
	Path     string
	Info     *os.FileInfo
}

type NanoStateIndex struct {
	stateRoots []string
	_id_index  map[string]int
	_fn_index  map[string]int
	_mt_index  map[int]NanoStateMeta
	_ct        int
}

func NewNanoStateIndex() *NanoStateIndex {
	nsf := new(NanoStateIndex)
	nsf.stateRoots = make([]string, 0)
	nsf._id_index = make(map[string]int)
	nsf._fn_index = make(map[string]int)
	nsf._mt_index = make(map[int]NanoStateMeta)

	return nsf
}

// AddStateRoot is used to chain-add another state root
func (nsf *NanoStateIndex) AddStateRoot(pth string) *NanoStateIndex {
	nsf.stateRoots = append(nsf.stateRoots, pth)
	return nsf
}

// AddStateRoots is used to chain-add another state roots (array)
func (nsf *NanoStateIndex) AddStateRoots(pth ...string) *NanoStateIndex {
	for _, p := range pth {
		nsf.AddStateRoot(p)
	}
	return nsf
}

// Index all the files in the all roots
func (nsf *NanoStateIndex) Index() {
	nsf._ct = len(nsf._mt_index)
	for _, root := range nsf.stateRoots {
		nsf.getPathFiles(root)
	}
}

// This only unmarshalls the state and fetches its ID
func (nsf *NanoStateIndex) getStateId(pth string) string {
	data, err := ioutil.ReadFile(pth)
	if err != nil {
		panic(err)
	}
	var state map[string]interface{}
	err = yaml.Unmarshal(data, &state)
	if err != nil {
		panic(err)
	}
	stateId, ex := state["id"]
	if !ex {
		panic("State has no id")
	}
	return stateId.(string)
}

func (nsf *NanoStateIndex) getPathFiles(root string) {
	err := filepath.Walk(root,
		func(pth string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				nsm := &NanoStateMeta{
					Id:       nsf.getStateId(pth),
					Filename: path.Base(pth),
					Path:     pth,
					Info:     &info,
				}
				nsf._mt_index[nsf._ct] = *nsm
				nsf._fn_index[nsm.Filename] = nsf._ct
				nsf._id_index[nsm.Id] = nsf._ct
				nsf._ct++
			}
			return nil
		})
	if err != nil {
		panic(err)
	}
}

func (nsf *NanoStateIndex) GetStateById(id string) *NanoStateMeta {
	fp, ok := nsf._id_index[id]
	if !ok {
		panic("ID does not exist")
	}
	nstm := nsf._mt_index[fp]
	return &nstm
}

func (nsf *NanoStateIndex) GetStateByFileName(name string) *NanoStateMeta {
	fp, ok := nsf._fn_index[name]
	if !ok {
		panic("Filename does not exist")
	}
	nstm := nsf._mt_index[fp]
	return &nstm
}
