// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package pkg

import (
	"fmt"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	fileConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/file"
	"github.com/rs/zerolog/log"
)

type Option func(gomod.Package, gomod.Module, *cdx.Component) error

func WithFiles(enabled bool) Option {
	return func(p gomod.Package, m gomod.Module, c *cdx.Component) error {
		if !enabled {
			return nil
		}

		var files []string
		files = append(files, p.GoFiles...)
		files = append(files, p.CgoFiles...)
		files = append(files, p.CFiles...)
		files = append(files, p.CXXFiles...)
		files = append(files, p.MFiles...)
		files = append(files, p.HFiles...)
		files = append(files, p.FFiles...)
		files = append(files, p.SFiles...)
		files = append(files, p.SwigFiles...)
		files = append(files, p.SwigCXXFiles...)
		files = append(files, p.SysoFiles...)
		files = append(files, p.EmbedFiles...)

		var fileComponents []cdx.Component

		for _, f := range files {
			fileComponent, err := fileConv.ToComponent(
				filepath.Join(p.Dir, f),
				f,
				fileConv.WithHashes(
					cdx.HashAlgoMD5,
					cdx.HashAlgoSHA1,
					cdx.HashAlgoSHA256,
					cdx.HashAlgoSHA384,
					cdx.HashAlgoSHA512,
				),
			)
			if err != nil {
				return err
			}

			fileComponents = append(fileComponents, *fileComponent)
		}

		if len(fileComponents) > 0 {
			c.Components = &fileComponents
		}

		return nil
	}
}

func ToComponent(p gomod.Package, m gomod.Module, options ...Option) (*cdx.Component, error) {
	log.Debug().
		Str("package", p.ImportPath).
		Msg("converting package to component")

	component := cdx.Component{
		Type:       cdx.ComponentTypeLibrary,
		Name:       p.ImportPath,
		Version:    m.Version,
		PackageURL: fmt.Sprintf("pkg:golang/%s@%s?type=package", p.ImportPath, m.Version),
	}

	for _, option := range options {
		if err := option(p, m, &component); err != nil {
			return nil, err
		}
	}

	return &component, nil
}