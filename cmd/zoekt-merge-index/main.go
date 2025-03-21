// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Command zoekt-merge-index merges a set of index shards into a compound shard.
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/zoekt/index"
)

// merge merges the input shards into a compound shard in dstDir. It returns the
// full path to the compound shard. The input shards are removed on success.
func merge(dstDir string, names []string) (string, error) {
	var files []index.IndexFile
	for _, fn := range names {
		f, err := os.Open(fn)
		if err != nil {
			return "", nil
		}
		defer f.Close()

		indexFile, err := index.NewIndexFile(f)
		if err != nil {
			return "", err
		}
		defer indexFile.Close()

		files = append(files, indexFile)
	}

	tmpName, dstName, err := index.Merge(dstDir, files...)
	if err != nil {
		return "", err
	}

	// Delete input shards.
	for _, name := range names {
		paths, err := index.IndexFilePaths(name)
		if err != nil {
			return "", fmt.Errorf("zoekt-merge-index: %w", err)
		}
		for _, p := range paths {
			if err := os.Remove(p); err != nil {
				return "", fmt.Errorf("zoekt-merge-index: failed to remove simple shard: %w", err)
			}
		}
	}

	// We only rename the compound shard if all simple shards could be deleted in the
	// previous step. This guarantees we won't have duplicate indexes.
	if err := os.Rename(tmpName, dstName); err != nil {
		return "", fmt.Errorf("zoekt-merge-index: failed to rename compound shard: %w", err)
	}

	return dstName, nil
}

func mergeCmd(paths []string) (string, error) {
	if paths[0] == "-" {
		paths = []string{}
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			paths = append(paths, strings.TrimSpace(scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			return "", err
		}
		log.Printf("merging %d paths from stdin", len(paths))
	}

	return merge(filepath.Dir(paths[0]), paths)
}

func explodeCmd(path string) error {
	return index.Explode(filepath.Dir(path), path)
}

func main() {
	switch subCommand := os.Args[1]; subCommand {
	case "merge":
		compoundShardPath, err := mergeCmd(os.Args[2:])
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(compoundShardPath)
	case "explode":
		if err := explodeCmd(os.Args[2]); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown subcommand %s", subCommand)
	}
}
