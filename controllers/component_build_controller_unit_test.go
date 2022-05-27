/*
Copyright 2021-2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestUpdateServiceAccountIfSecretNotLinked(t *testing.T) {
	type args struct {
		gitSecretName  string
		serviceAccount *corev1.ServiceAccount
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "present",
			args: args{
				gitSecretName: "present",
				serviceAccount: &corev1.ServiceAccount{
					Secrets: []corev1.ObjectReference{
						{
							Name: "present",
						},
					},
				},
			},
			want: false, // since it was present, this implies the SA wasn't updated.
		},
		{
			name: "not present",
			args: args{
				gitSecretName: "not-present",
				serviceAccount: &corev1.ServiceAccount{
					Secrets: []corev1.ObjectReference{
						{
							Name: "something-else",
						},
					},
				},
			},
			want: true, // since it wasn't present, this implies the SA was updated.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := updateServiceAccountIfSecretNotLinked(tt.args.gitSecretName, tt.args.serviceAccount); got != tt.want {
				t.Errorf("UpdateServiceAccountIfSecretNotLinked() = %v, want %v", got, tt.want)
			}
		})
	}
}
