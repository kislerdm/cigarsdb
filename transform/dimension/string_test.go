package dimension

import "testing"

func Test_toCapFirstLetter(t *testing.T) {
	tests := []struct {
		s    string
		want string
	}{
		{
			s:    "foo ár",
			want: "Foo Ár",
		},
		{
			s:    "foo    ár",
			want: "Foo    Ár",
		},
	}
	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			if got := toCapFirstLetters(tt.s); got != tt.want {
				t.Errorf("toCapFirstLetters() = %v, want %v", got, tt.want)
			}
		})
	}
}
