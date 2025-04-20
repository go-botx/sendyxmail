package mutemanager

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type MuteManager struct {
	file         string
	mutedEntries map[string]bool
	mutex        sync.RWMutex
}

func New(file string) (*MuteManager, error) {
	var err error
	mm := &MuteManager{
		mutedEntries: map[string]bool{},
		file:         file,
	}
	mm.file, err = filepath.Abs(mm.file)
	if err != nil {
		return nil, err
	}
	err = mm.loadFile()
	if err != nil {
		return nil, err
	}
	return mm, err
}

func (mm *MuteManager) GetMute(entry string) bool {
	var state bool
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	_, state = mm.mutedEntries[entry]
	return state
}

func (mm *MuteManager) SetMute(entry string, state bool) (changed bool, err error) {
	mm.mutex.RLock()
	_, currentState := mm.mutedEntries[entry]
	mm.mutex.RUnlock()
	if currentState == state {
		return false, nil
	}
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// Check again
	_, currentState = mm.mutedEntries[entry]
	if currentState == state {
		return false, nil
	}

	if state {
		mm.mutedEntries[entry] = true
	} else {
		delete(mm.mutedEntries, entry)
	}

	err = mm.saveFile()
	if err != nil {
		// Roll back
		if state {
			delete(mm.mutedEntries, entry)
		} else {
			mm.mutedEntries[entry] = true
		}
		return false, err
	}
	return true, nil
}

func (mm *MuteManager) saveFile() error {

	err := mm.createCopy(mm.file, mm.file+".mmbak")
	if err != nil {
		return err
	}
	tempFile, err := os.CreateTemp(filepath.Dir(mm.file), filepath.Base(mm.file)+".*.mmtemp")
	if err != nil {
		return err
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	writer := bufio.NewWriter(tempFile)

	for k := range mm.mutedEntries {
		_, err = writer.WriteString(k + "\n")
		if err != nil {
			return err
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	err = tempFile.Sync()
	if err != nil {
		return err
	}
	err = tempFile.Close()
	if err != nil {
		return err
	}
	err = mm.createCopy(tempFile.Name(), mm.file)
	if err != nil {
		return err
	}
	return nil
}

func (mm *MuteManager) loadFile() error {
	var file *os.File
	var err error
	file, err = os.Open(mm.file)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			err = mm.saveFile()
			if err != nil {
				return err
			}
			file, err = os.Open(mm.file)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lines := map[string]bool{}
	for scanner.Scan() {
		lines[scanner.Text()] = true
	}
	err = scanner.Err()
	if err != nil {
		return err
	}
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	for k := range mm.mutedEntries {
		delete(mm.mutedEntries, k)
	}
	for k := range lines {
		mm.mutedEntries[k] = true
	}
	return nil
}

func (mm *MuteManager) createCopy(src, dst string) error {
	if !filepath.IsAbs(src) || !filepath.IsAbs(dst) {
		return fmt.Errorf("path names MUST be absolute")
	}
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	err = destFile.Sync()
	return err
}
