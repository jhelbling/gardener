// Copyright 2018 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package imagevector_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/utils/pointer"

	. "github.com/gardener/gardener/pkg/utils/imagevector"
	"github.com/gardener/gardener/pkg/utils/test"
)

var _ = Describe("imagevector", func() {
	Describe("> ImageVector", func() {
		var (
			image1Src1Vector     ImageVector
			image1Src1VectorJSON string
			image1Src1VectorYAML string

			arm64 = "arm64"
			amd64 = "amd64"

			k8s164                         = "1.6.4"
			k8s164WithSuffix               = "1.6.4-foo.5"
			k8s180                         = "1.8.0"
			k8s113                         = "1.13"
			k8s1142                        = "1.14.2"
			k8s114x                        = "1.14.x"
			k8s1170                        = "1.17.0"
			k8s164RuntimeVersion           = RuntimeVersion(k8s164)
			k8s164WithSuffixRuntimeVersion = RuntimeVersion(k8s164WithSuffix)
			k8s164TargetVersion            = TargetVersion(k8s164)
			k8s1170TargetVersion           = TargetVersion(k8s1170)
			k8s180RuntimeVersion           = RuntimeVersion(k8s180)
			k8s180TargetVersion            = TargetVersion(k8s180)
			k8s113TargetVersion            = TargetVersion(k8s113)
			k8s1142TargetVersion           = TargetVersion(k8s1142)

			tag1, tag2, tag3, tag4, tag5, tag6, tag7 string
			repo1, repo2, repo3, repo4               string

			greaterEquals16Smaller18, greaterEquals18, equals117 string

			image1Name                                                             string
			image1Src1, image1Src2, image1Src3, image1Src4, image1Src5, image1Src6 *ImageSource
			image1Src1Arm64, image1Src5Arm64, image1Src5Wildcard                   *ImageSource

			image2Name string
			image2Src1 *ImageSource

			image3Name string
			image3Src1 *ImageSource

			image4Name                                                                          string
			image4Src1, image4Src2, image4Src3, image4Src4, image4Src5, image4Src6, image4Src7  *ImageSource
			image4Src1Arm64, image4Src4Arm64, image4Src5Arm64, image4Src6Arm64, image4Src7Arm64 *ImageSource
		)

		resetValues := func() {
			k8s164 = "1.6.4"
			k8s180 = "1.8.0"
			k8s164RuntimeVersion = RuntimeVersion(k8s164)
			k8s164TargetVersion = TargetVersion(k8s164)
			k8s180RuntimeVersion = RuntimeVersion(k8s180)
			k8s180TargetVersion = TargetVersion(k8s180)

			tag1 = "tag1"
			tag2 = "tag2"
			tag3 = "tag3"
			tag4 = "tag4"
			tag5 = "tag5"
			tag6 = "tag6"
			tag7 = "tag7"

			repo1 = "repo1"
			repo2 = "repo2"
			repo3 = "repo3"
			repo3 = "repo4"

			greaterEquals16Smaller18 = ">= 1.6, < 1.8"
			greaterEquals18 = ">= 1.8"
			equals117 = "= 1.17.0"

			image1Name = "image1"
			image1Src1 = &ImageSource{
				Name:           image1Name,
				Repository:     repo1,
				Tag:            &tag1,
				RuntimeVersion: &greaterEquals16Smaller18,
			}
			image1Src1Arm64 = &ImageSource{
				Name:           image1Name,
				Repository:     repo1,
				Tag:            &tag1,
				RuntimeVersion: &greaterEquals16Smaller18,
				Architectures:  []string{arm64},
			}
			image1Src2 = &ImageSource{
				Name:           image1Name,
				Repository:     repo1,
				Tag:            &tag1,
				RuntimeVersion: &greaterEquals18,
			}
			image1Src3 = &ImageSource{
				Name:           image1Name,
				Repository:     repo2,
				Tag:            &tag1,
				RuntimeVersion: &greaterEquals16Smaller18,
			}
			image1Src4 = &ImageSource{
				Name:           image1Name,
				Repository:     repo1,
				Tag:            &tag2,
				RuntimeVersion: &greaterEquals16Smaller18,
			}
			image1Src5 = &ImageSource{
				Name:       image1Name,
				Repository: repo1,
				Tag:        &tag1,
			}
			image1Src5Arm64 = &ImageSource{
				Name:          image1Name,
				Repository:    repo1,
				Tag:           &tag1,
				Architectures: []string{arm64},
			}
			image1Src5Wildcard = &ImageSource{
				Name:          image1Name,
				Repository:    repo1,
				Tag:           &tag1,
				Architectures: []string{arm64, amd64},
			}
			image1Src6 = &ImageSource{
				Name:           image1Name,
				Repository:     repo1,
				RuntimeVersion: &greaterEquals16Smaller18,
			}

			image2Name = "image2"
			image2Src1 = &ImageSource{
				Name:           image2Name,
				Repository:     repo2,
				Tag:            &tag2,
				RuntimeVersion: &greaterEquals16Smaller18,
			}

			image3Name = "image3"
			image3Src1 = &ImageSource{
				Name:       image3Name,
				Repository: repo3,
			}

			image4Name = "image4"
			image4Src1 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag1,
				TargetVersion: &greaterEquals16Smaller18,
			}
			image4Src1Arm64 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag1,
				TargetVersion: &greaterEquals16Smaller18,
				Architectures: []string{arm64},
			}
			image4Src2 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag2,
				TargetVersion: &k8s113,
			}
			image4Src3 = &ImageSource{
				Name:       image4Name,
				Repository: repo4,
				Tag:        &tag3,
			}
			image4Src4 = &ImageSource{
				Name:           image4Name,
				Repository:     repo4,
				Tag:            &tag4,
				RuntimeVersion: &greaterEquals16Smaller18,
				TargetVersion:  &k8s113,
			}
			image4Src4Arm64 = &ImageSource{
				Name:           image4Name,
				Repository:     repo4,
				Tag:            &tag4,
				RuntimeVersion: &greaterEquals16Smaller18,
				TargetVersion:  &k8s113,
				Architectures:  []string{arm64},
			}
			image4Src5 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag5,
				TargetVersion: &equals117,
			}
			image4Src5Arm64 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag5,
				TargetVersion: &equals117,
				Architectures: []string{arm64},
			}
			image4Src6 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag6,
				TargetVersion: &k8s114x,
			}
			image4Src6Arm64 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag6,
				TargetVersion: &k8s114x,
				Architectures: []string{arm64},
			}
			image4Src7 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag7,
				TargetVersion: &k8s1142,
			}
			image4Src7Arm64 = &ImageSource{
				Name:          image4Name,
				Repository:    repo4,
				Tag:           &tag7,
				TargetVersion: &k8s1142,
				Architectures: []string{arm64},
			}

			image1Src1Vector = ImageVector{image1Src1}

			image1Src1VectorJSON = fmt.Sprintf(`
{
	"images": [
		{
			"name": "%s",
			"repository": "%s",
			"tag": "%s",
			"runtimeVersion": "%s"
		}
	]
}`, image1Src1.Name, image1Src1.Repository, *image1Src1.Tag, *image1Src1.RuntimeVersion)

			image1Src1VectorYAML = fmt.Sprintf(`
images:
  - name: "%s"
    repository: "%s"
    tag: "%s"
    runtimeVersion: "%s"`, image1Src1.Name, image1Src1.Repository, *image1Src1.Tag, *image1Src1.RuntimeVersion)
		}
		resetValues()
		BeforeEach(resetValues)

		Describe("#Read", func() {
			It("should successfully read a JSON image vector", func() {
				vector, err := Read([]byte(image1Src1VectorJSON))
				Expect(err).NotTo(HaveOccurred())
				Expect(vector).To(Equal(image1Src1Vector))
			})

			It("should successfully read a YAML image vector", func() {
				vector, err := Read([]byte(image1Src1VectorYAML))
				Expect(err).NotTo(HaveOccurred())
				Expect(vector).To(Equal(image1Src1Vector))
			})
		})

		Describe("#ReadFile", func() {
			It("should successfully read the file and close it", func() {
				tmpFile, cleanup := withTempFile("imagevector", []byte(image1Src1VectorJSON))
				defer cleanup()

				vector, err := ReadFile(tmpFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(vector).To(Equal(image1Src1Vector))
			})
		})

		DescribeTable("#Merge",
			func(v1, v2, expected ImageVector) {
				Expect(Merge(v1, v2)).To(Equal(expected))
			},

			Entry("no override", ImageVector{image1Src1}, ImageVector{image1Src2}, ImageVector{image1Src1, image1Src2}),
			Entry("one override, one addition", ImageVector{image1Src1, image2Src1}, ImageVector{image1Src3, image3Src1}, ImageVector{image1Src3, image2Src1, image3Src1}),
			Entry("tag is kept", ImageVector{image1Src1}, ImageVector{image1Src6}, ImageVector{image1Src1}),
			Entry("tag override", ImageVector{image1Src1}, ImageVector{image1Src4}, ImageVector{image1Src4}),
		)

		Describe("#WithEnvOverride", func() {
			It("should override the ImageVector with the settings of the env variable", func() {
				var (
					vector = ImageVector{image1Src3, image2Src1}
				)

				file, cleanup := withTempFile("imagevector", []byte(image1Src1VectorJSON))

				defer cleanup()
				defer test.WithEnvVar(OverrideEnv, file.Name())()

				Expect(WithEnvOverride(vector)).To(Equal(ImageVector{image1Src1, image2Src1}))
			})

			It("should keep the vector as-is if the env variable is not set", func() {
				Expect(WithEnvOverride(image1Src1Vector)).To(Equal(image1Src1Vector))
			})
		})

		DescribeTable("#FindImage",
			func(vec ImageVector, name string, opts []FindOptionFunc, imageMatcher, errorMatcher types.GomegaMatcher) {
				image, err := vec.FindImage(name, opts...)

				Expect(image).To(imageMatcher)
				Expect(err).To(errorMatcher)
			},

			Entry("no entries, no match",
				ImageVector{},
				image1Name, nil,
				BeNil(), HaveOccurred()),

			Entry("single entry, match with runtime wildcard",
				ImageVector{image1Src1},
				image1Name, nil,
				Equal(image1Src1.ToImage(nil)), Not(HaveOccurred())),
			Entry("single entry w/ arch, no arch opts, no runtime version opts",
				ImageVector{image1Src1Arm64},
				image1Name, nil,
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, no runtime version opts",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{Architecture(arm64)},
				Equal(image1Src1.ToImage(nil)), Not(HaveOccurred())),

			Entry("single entry, match with runtime version",
				ImageVector{image1Src1},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion},
				Equal(image1Src1.ToImage(nil)), Not(HaveOccurred())),
			Entry("single entry w/ arch, no arch opts, match with runtime version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion},
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, match with runtime version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, Architecture(arm64)},
				Equal(image1Src1.ToImage(nil)), Not(HaveOccurred())),

			Entry("single entry w/ suffix, match with runtime version",
				ImageVector{image1Src1},
				image1Name, []FindOptionFunc{k8s164WithSuffixRuntimeVersion},
				Equal(image1Src1.ToImage(nil)), Not(HaveOccurred())),
			Entry("single entry w/ arch and suffix, no arch opts, match with runtime version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164WithSuffixRuntimeVersion},
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch and suffix, arch opts, match with runtime version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164WithSuffixRuntimeVersion, Architecture(arm64)},
				Equal(image1Src1.ToImage(nil)), Not(HaveOccurred())),

			Entry("single entry, match with runtime and target version",
				ImageVector{image1Src1},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, k8s164TargetVersion},
				Equal(image1Src1.ToImage(&k8s164)), Not(HaveOccurred())),
			Entry("single entry w/ arch, no arch opts, match with runtime and target version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, k8s164TargetVersion},
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, match with runtime and target version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, k8s164TargetVersion, Architecture(arm64)},
				Equal(image1Src1.ToImage(&k8s164)), Not(HaveOccurred())),

			Entry("single entry, match with runtime and non-runtime target version",
				ImageVector{image1Src1},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, k8s180TargetVersion},
				Equal(image1Src1.ToImage(&k8s180)), Not(HaveOccurred())),
			Entry("single entry w/ arch, no arch opts, match with runtime and non-runtime target version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, k8s180TargetVersion},
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, match with runtime and non-runtime target version",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, k8s180TargetVersion, Architecture(arm64)},
				Equal(image1Src1.ToImage(&k8s180)), Not(HaveOccurred())),

			Entry("single entry, name mismatch",
				ImageVector{image1Src1},
				image2Name, nil,
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, no arch opts, name mismatch",
				ImageVector{image1Src1Arm64},
				image2Name, nil,
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, name mismatch",
				ImageVector{image1Src1Arm64},
				image2Name, []FindOptionFunc{Architecture(arm64)},
				BeNil(), HaveOccurred()),

			Entry("single entry, runtime version mismatch",
				ImageVector{image1Src1},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion},
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, no arch opts, runtime version mismatch",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion},
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, runtime version mismatch",
				ImageVector{image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion, Architecture(arm64)},
				BeNil(), HaveOccurred()),

			Entry("single entry, no runtime version, match with runtime wildcard",
				ImageVector{image1Src5},
				image1Name, nil,
				Equal(image1Src5.ToImage(nil)), Not(HaveOccurred())),
			Entry("single entry w/ arch, no arch opts, no runtime version, match with runtime wildcard",
				ImageVector{image1Src5Arm64},
				image1Name, nil,
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, no runtime version, match with runtime wildcard",
				ImageVector{image1Src5Arm64},
				image1Name, []FindOptionFunc{Architecture(arm64)},
				Equal(image1Src5.ToImage(nil)), Not(HaveOccurred())),

			Entry("single entry, no runtime version, match with runtime version",
				ImageVector{image1Src5},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion},
				Equal(image1Src5.ToImage(nil)), Not(HaveOccurred())),
			Entry("single entry w/ arch, no arch opts, no runtime version, match with runtime version",
				ImageVector{image1Src5Arm64},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion},
				BeNil(), HaveOccurred()),
			Entry("single entry w/ arch, arch opts, no runtime version, match with runtime version",
				ImageVector{image1Src5Arm64},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion, Architecture(arm64)},
				Equal(image1Src5.ToImage(nil)), Not(HaveOccurred())),

			Entry("two entries + one entry w/ arch, no arch opts, match with runtime wildcard",
				ImageVector{image1Src5, image1Src1, image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion},
				Equal(image1Src5.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + one entry w/ arch, arch opts, match with runtime wildcard",
				ImageVector{image1Src5, image1Src1, image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion, Architecture(arm64)},
				Equal(image1Src5.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + one entry w/ arch, arch opts (wildcard), match with runtime wildcard",
				ImageVector{image1Src5, image1Src1, image1Src1Arm64, image1Src5Wildcard},
				image1Name, []FindOptionFunc{k8s180RuntimeVersion, Architecture(arm64)},
				Equal(image1Src5Wildcard.ToImage(nil)), Not(HaveOccurred())),

			Entry("two entries + one entry w/ arch, no arch opts, match with runtime version",
				ImageVector{image1Src5, image1Src1, image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion},
				Equal(image1Src1.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + one entry w/ arch, arch opts, match with runtime version",
				ImageVector{image1Src5, image1Src1, image1Src1Arm64},
				image1Name, []FindOptionFunc{k8s164RuntimeVersion, Architecture(arm64)},
				Equal(image1Src1Arm64.ToImage(nil)), Not(HaveOccurred())),

			Entry("two entries + one entry w/ arch, no arch opts, no runtime version, match with target version",
				ImageVector{image4Src1, image4Src2, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s113TargetVersion},
				Equal(image4Src2.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + one entry w/ arch, arch opts, no runtime version, match with runtime wildcard",
				ImageVector{image4Src1, image4Src2, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s113TargetVersion, Architecture(arm64)},
				Equal(image4Src2.ToImage(nil)), Not(HaveOccurred())),

			Entry("two entries + one entry w/ arch, no arch opts, runtime and target version, no match",
				ImageVector{image4Src1, image4Src4, image1Src1Arm64},
				image4Name, []FindOptionFunc{k8s180RuntimeVersion, k8s113TargetVersion},
				BeNil(), HaveOccurred()),
			Entry("two entries + one entry w/ arch, arch opts, runtime and target version, no match",
				ImageVector{image4Src1, image4Src4, image1Src1Arm64},
				image4Name, []FindOptionFunc{k8s180RuntimeVersion, k8s113TargetVersion, Architecture(arm64)},
				BeNil(), HaveOccurred()),

			Entry("two entries + one entry w/ arch, no arch opts, runtime and target version, no match",
				ImageVector{image4Src1, image4Src4, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s164RuntimeVersion, k8s1142TargetVersion},
				BeNil(), HaveOccurred()),
			Entry("two entries + one entry w/ arch, arch opts, runtime and target version, no match",
				ImageVector{image4Src1, image4Src4, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s164RuntimeVersion, k8s1142TargetVersion, Architecture(arm64)},
				BeNil(), HaveOccurred()),

			Entry("two entries + one entry w/ arch, no arch opts, runtime and target version, match with both",
				ImageVector{image4Src1, image4Src4, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s164RuntimeVersion, k8s113TargetVersion},
				Equal(image4Src4.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + one entry w/ arch, arch opts, runtime and target version, match with runtime wildcard",
				ImageVector{image4Src1, image4Src4, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s164RuntimeVersion, k8s113TargetVersion, Architecture(arm64)},
				Equal(image4Src4.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + two entries w/ arch, arch opts, runtime and target version, match with both",
				ImageVector{image4Src1, image4Src4, image4Src4Arm64},
				image4Name, []FindOptionFunc{k8s164RuntimeVersion, k8s113TargetVersion, Architecture(arm64)},
				Equal(image4Src4Arm64.ToImage(nil)), Not(HaveOccurred())),

			Entry("two entries + one entry w/ arch, no arch opts, no runtime version, exact match with target version",
				ImageVector{image4Src6, image4Src7, image4Src6Arm64},
				image4Name, []FindOptionFunc{k8s1142TargetVersion},
				Equal(image4Src7.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + one entry w/ arch, arch opts, no runtime version, exact match with target version",
				ImageVector{image4Src6, image4Src7, image4Src6Arm64},
				image4Name, []FindOptionFunc{k8s1142TargetVersion, Architecture(arm64)},
				Equal(image4Src7.ToImage(nil)), Not(HaveOccurred())),
			Entry("two entries + two entries w/ arch, arch opts, no runtime version, exact match with target version",
				ImageVector{image4Src6, image4Src7, image4Src6Arm64, image4Src7Arm64},
				image4Name, []FindOptionFunc{k8s1142TargetVersion, Architecture(arm64)},
				Equal(image4Src7Arm64.ToImage(nil)), Not(HaveOccurred())),

			Entry("three entries + one entry w/ arch, no arch opts, no runtime version, match with target version",
				ImageVector{image4Src1, image4Src2, image4Src3, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s113TargetVersion},
				Equal(image4Src2.ToImage(nil)), Not(HaveOccurred())),
			Entry("three entries + one entry w/ arch, arch opts, no runtime version, match with runtime wildcard",
				ImageVector{image4Src1, image4Src2, image4Src3, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s113TargetVersion, Architecture(arm64)},
				Equal(image4Src2.ToImage(nil)), Not(HaveOccurred())),

			Entry("three entries + one entry w/ arch, no arch opts, no runtime version, match with target version",
				ImageVector{image4Src1, image4Src2, image4Src3, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s1142TargetVersion},
				Equal(image4Src3.ToImage(nil)), Not(HaveOccurred())),
			Entry("three entries + one entry w/ arch, arch opts, no runtime version, match with runtime wildcard",
				ImageVector{image4Src1, image4Src2, image4Src3, image4Src1Arm64},
				image4Name, []FindOptionFunc{k8s1142TargetVersion, Architecture(arm64)},
				Equal(image4Src3.ToImage(nil)), Not(HaveOccurred())),
			Entry("three entries + two entries w/ arch, arch opts, no runtime version, match with target version",
				ImageVector{image4Src1, image4Src2, image4Src3, image4Src1Arm64, image4Src7Arm64},
				image4Name, []FindOptionFunc{k8s1142TargetVersion, Architecture(arm64)},
				Equal(image4Src7Arm64.ToImage(nil)), Not(HaveOccurred())),

			Entry("four entries + two entries w/ arch, no arch opts, runtime and target version, match with both, prio equal match",
				ImageVector{image4Src1, image4Src2, image4Src3, image4Src5, image4Src1Arm64, image4Src5Arm64},
				image4Name, []FindOptionFunc{k8s1170TargetVersion},
				Equal(image4Src5.ToImage(nil)), Not(HaveOccurred())),
			Entry("four entries + two entries w/ arch, arch opts, runtime and target version, match with both, prio equal match",
				ImageVector{image4Src1, image4Src2, image4Src3, image4Src5, image4Src1Arm64, image4Src5Arm64},
				image4Name, []FindOptionFunc{k8s1170TargetVersion, Architecture(arm64)},
				Equal(image4Src5Arm64.ToImage(nil)), Not(HaveOccurred())),
		)

		Describe("#FindImages", func() {
			It("should collect the found images in a map", func() {
				v := ImageVector{image1Src1, image2Src1}

				images, err := FindImages(v, []string{image1Name, image2Name})

				Expect(err).NotTo(HaveOccurred())
				Expect(images).To(Equal(map[string]*Image{
					image1Name: image1Src1.ToImage(nil),
					image2Name: image2Src1.ToImage(nil),
				}))
			})

			It("should error if it couldn't find an image", func() {
				v := ImageVector{image1Src1, image2Src1}

				_, err := FindImages(v, []string{image1Name, image2Name, image3Name})
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("> Image", func() {
		Describe("#String", func() {
			It("should return the string representation of the image (w/o normal tag)", func() {
				repo := "my-repo"

				image := Image{
					Name:       "my-image",
					Repository: repo,
				}

				Expect(image.String()).To(Equal(repo))
			})

			It("should return the string representation of the image (w/ normal tag)", func() {
				var (
					repo = "my-repo"
					tag  = "1.2.3"
				)

				image := Image{
					Name:       "my-image",
					Repository: repo,
					Tag:        &tag,
				}

				Expect(image.String()).To(Equal(fmt.Sprintf("%s:%s", repo, tag)))
			})

			It("should return the string representation of the image (w/ sha256 tag)", func() {
				var (
					repo = "my-repo"
					tag  = "sha256:fooooooo0oooobar"
				)

				image := Image{
					Name:       "my-image",
					Repository: repo,
					Tag:        &tag,
				}

				Expect(image.String()).To(Equal(repo + "@" + tag))
			})
		})
	})

	Describe("> ImageSource", func() {
		Describe("#ToImage", func() {
			It("should return an image with the same tag", func() {
				var (
					name       = "foo"
					repository = "repo"
					tag        = "v1"

					source = ImageSource{
						Name:       name,
						Repository: repository,
						Tag:        &tag,
					}
				)

				image := source.ToImage(pointer.String("1.8.0"))

				Expect(image).To(Equal(&Image{
					Name:       name,
					Repository: repository,
					Tag:        &tag,
				}))
			})

			It("should return an image with the given version as tag", func() {
				var (
					name       = "foo"
					repository = "repo"
					version    = "1.8.0"

					source = ImageSource{
						Name:       name,
						Repository: repository,
					}
				)

				image := source.ToImage(&version)

				Expect(image).To(Equal(&Image{
					Name:       name,
					Repository: repository,
					Tag:        pointer.String(fmt.Sprintf("v%s", version)),
				}))
			})
		})
	})

	Describe("#ImageMapToValues", func() {
		It("should return the expected map", func() {
			var (
				image1Key        = "foo"
				image1Name       = "baz"
				image1Repository = "baz"
				image1Tag        = "barbaz"

				image2Key        = "bar"
				image2Name       = "baz"
				image2Repository = "foo"

				imageMap = map[string]*Image{
					image1Key: {
						Name:       image1Name,
						Repository: image1Repository,
						Tag:        &image1Tag,
					},
					image2Key: {
						Name:       image2Name,
						Repository: image2Repository,
					},
				}
			)

			Expect(ImageMapToValues(imageMap)).To(Equal(map[string]interface{}{
				image1Key: image1Name + ":" + image1Tag,
				image2Key: image2Repository,
			}))
		})
	})
})

func withTempFile(pattern string, data []byte) (*os.File, func()) {
	tmpFile, err := os.CreateTemp("", pattern)
	Expect(err).NotTo(HaveOccurred())
	Expect(os.WriteFile(tmpFile.Name(), data, os.ModePerm)).To(Succeed())

	return tmpFile, func() {
		if err := tmpFile.Close(); err != nil {
			GinkgoT().Logf("Could not close temp file %q: %v", tmpFile.Name(), err)
		}
		if err := os.Remove(tmpFile.Name()); err != nil {
			GinkgoT().Logf("Could not delete temp file %q: %v", tmpFile.Name(), err)
		}
	}
}
