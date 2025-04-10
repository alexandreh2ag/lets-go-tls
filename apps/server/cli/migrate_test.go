package cli

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"io"
	"path/filepath"
	"testing"
)

func TestGetMigrateRunFn_SuccessTraefik(t *testing.T) {
	acmePath := "/app/letsencrypt/acme.json"
	outputPath := "/app/state.json"
	ctx := context.TestContext(nil)
	fs := afero.NewMemMapFs()
	ctx.Fs = fs
	_ = fs.MkdirAll(filepath.Base(acmePath), 0755)
	_ = afero.WriteFile(fs, acmePath, []byte("{}"), 0755)
	want := []byte("{\"account\":{\"Email\":\"\",\"Registration\":null,\"Key\":null},\"certificates\":[]}")
	viper.Reset()
	viper.SetFs(ctx.Fs)
	cmd := GetMigrateCmd(ctx)
	cmd.SetArgs([]string{
		"--type", "traefik",
		"--path", acmePath,
		"--output", outputPath,
	})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	assert.NoError(t, err)
	outputRaw, errRead := afero.ReadFile(fs, outputPath)
	assert.NoError(t, errRead)
	assert.Equal(t, string(want), string(outputRaw))
}

func TestGetMigrateRunFn_SuccessCertbot(t *testing.T) {
	acmeDir := "/app/letsencrypt"
	outputPath := "/app/state.json"
	ctx := context.TestContext(nil)

	regrValid := "{\"body\": {}, \"uri\": \"http://server/acme/name/directory/azerty\"}"
	keyValid := "{\"n\": \"9pbYlW2rfzxM9BnhKnWXWQvOaSofpbgpLnDZqBb5UJuOexknabPVVjwAJCXyn004Q4J7mjkNsc6xK42Pd8l_2Z1FODh7NJjw_TDBIMn44WQqRF1tZe29WIZBvaW1g7VXkJ4oI93osUf4sqQ-24tUOXYktAVirITk4tIgfiKAV-KI28Cye2CkCmMZtWi_yKoySo-dVLUj1BfZ3BBeDRNVEr2VcNtq6qqZ54phtblcRJPaJ2ZRr77hZJN-DU5Q1IttnJqf1OJkndlRbAdgZ5uHH_k3w-6MlobD1YZPT6bSsSoZMx6YRt14PVWLhuGu2NB_l3JFbGHF2kw3JsQUX5695w\", \"e\": \"AQAB\", \"d\": \"Cwc2CoDMIbWdr3EU2-REb4wSoCttHOk-lkAfy9-zKMK8PY8jbxvG18N5MiSsVCmM8Z_9TSluDlyzGcKL_wx49k-NT_VetTx90qUCjifaRKJJLaOMB_n49OOUPxoRIyHSF7qrCueax4rzeXHVCdaSXruE6rQu9I5A-k_xehKq8MMvaDBSY_Zwbmyd3QkeXSZq08CRL0nmKizwKamoEbxNdeHZDXyYtJKUaFRpS311Kr74OuC6C7jQRSIz4GlAN2DhgLMnJaxFZkvIWknkprTHRzlmCj6RNmRF9OVFlBhJTJdJjFAyfMLpdwzOki-Q5YdbxWrcA-7NJgHvvcZf1k_leQ\", \"p\": \"_dc0cSGPqeGwXUs1zVuOQc7sJODIjlSZ-Jt8GJ1geQBU8wLCFYkw-Fa6EfJUxk8qhgHpjgelLCv1m8x5fxPeUVafWSug7XqnwSvDpaThS1SNbzOPsyYYSI7RBuVLJz5y949fxziPawf0-Ip6Vv1zKaGKbz-gZFyJyZ8wxRbIaqU\", \"q\": \"-K_ZgKYrFHX96ZnjsVUyTQ0HZUAJGvNMaicA8sa_EgiDGtiDLO0yn6btV4uaVMiXy_5GRPikaHO0xD3DYilkGk0WGnswmhy24NyV7H5NDqNyT6BNFDldEe0_8N3WFs1ppY-tzWQgnSbazIuLq5wbR2TcyzAXJM4bXTjIXD0RvJs\", \"dp\": \"X_-QXKyVmWi5-z5jVYunjCnGUtgT4QgLxuJ-lwyFnJ1Mgy7q2Zwtwe2CpyDfgLQV3-I_QrCPfdoOI-K7nByWemppDx8Z2FYHtV1ou29UqNmMN57nXJgunNouIQ58UpYigo9daIlya8bxlUFBrT5i3D60jQgiD3KTxYWHuzG3j-U\", \"dq\": \"vqvVT9YX4bA15g2fji-MSZdbvD3EvW0dgaY2C-5mQSVVoBexO5kB33bVMvJOegLyn-1tYyRgqdVNG8lfLLwfjCreb59UPFUXwoBeUtzjp_2Zf4OglYkF2uDUzZDIvOLxxYbL81Z5ywQnbfzwMbuXDr4-q2tL02RThr4qoM4R81E\", \"qi\": \"LcHVIWfnbkq668LDsqJk48fsNwfJUjmxsvUUrf9YYVrjU2v4r6SV0DYb-UWjwccc5KZpnNY1sZGnbpSMoob3fqDhPb50cLMgLfj7DmVxR9hGIRnNgAYAf2f3mUfAyvjVwJGIKGoBS8vsnXbDmzZRGntKBSRIFvRYkCkVDEoXi8Q\", \"kty\": \"RSA\"}"

	fs := afero.NewMemMapFs()
	ctx.Fs = fs
	_ = fs.MkdirAll(filepath.Base(acmeDir), 0755)
	_ = fs.MkdirAll(filepath.Join(acmeDir, "live"), 0755)
	_ = fs.MkdirAll(filepath.Join(acmeDir, "accounts"), 0755)
	accountDir := filepath.Join(acmeDir, "accounts", "server", "acme", "name", "directory", "azerty")
	_ = fs.MkdirAll(accountDir, 0755)
	_ = afero.WriteFile(fs, filepath.Join(accountDir, "regr.json"), []byte(regrValid), 0644)
	_ = afero.WriteFile(fs, filepath.Join(accountDir, "private_key.json"), []byte(keyValid), 0644)
	want := []byte("{\"account\":{\"Email\":\"\",\"Registration\":{\"body\":{},\"uri\":\"http://server/acme/name/directory/azerty\"},\"Key\":\"MIIEowIBAAKCAQEA9pbYlW2rfzxM9BnhKnWXWQvOaSofpbgpLnDZqBb5UJuOexknabPVVjwAJCXyn004Q4J7mjkNsc6xK42Pd8l/2Z1FODh7NJjw/TDBIMn44WQqRF1tZe29WIZBvaW1g7VXkJ4oI93osUf4sqQ+24tUOXYktAVirITk4tIgfiKAV+KI28Cye2CkCmMZtWi/yKoySo+dVLUj1BfZ3BBeDRNVEr2VcNtq6qqZ54phtblcRJPaJ2ZRr77hZJN+DU5Q1IttnJqf1OJkndlRbAdgZ5uHH/k3w+6MlobD1YZPT6bSsSoZMx6YRt14PVWLhuGu2NB/l3JFbGHF2kw3JsQUX5695wIDAQABAoIBAAsHNgqAzCG1na9xFNvkRG+MEqArbRzpPpZAH8vfsyjCvD2PI28bxtfDeTIkrFQpjPGf/U0pbg5csxnCi/8MePZPjU/1XrU8fdKlAo4n2kSiSS2jjAf5+PTjlD8aESMh0he6qwrnmseK83lx1QnWkl67hOq0LvSOQPpP8XoSqvDDL2gwUmP2cG5snd0JHl0matPAkS9J5ios8CmpqBG8TXXh2Q18mLSSlGhUaUt9dSq++Drgugu40EUiM+BpQDdg4YCzJyWsRWZLyFpJ5Ka0x0c5Zgo+kTZkRfTlRZQYSUyXSYxQMnzC6XcMzpIvkOWHW8Vq3APuzSYB773GX9ZP5XkCgYEA/dc0cSGPqeGwXUs1zVuOQc7sJODIjlSZ+Jt8GJ1geQBU8wLCFYkw+Fa6EfJUxk8qhgHpjgelLCv1m8x5fxPeUVafWSug7XqnwSvDpaThS1SNbzOPsyYYSI7RBuVLJz5y949fxziPawf0+Ip6Vv1zKaGKbz+gZFyJyZ8wxRbIaqUCgYEA+K/ZgKYrFHX96ZnjsVUyTQ0HZUAJGvNMaicA8sa/EgiDGtiDLO0yn6btV4uaVMiXy/5GRPikaHO0xD3DYilkGk0WGnswmhy24NyV7H5NDqNyT6BNFDldEe0/8N3WFs1ppY+tzWQgnSbazIuLq5wbR2TcyzAXJM4bXTjIXD0RvJsCgYBf/5BcrJWZaLn7PmNVi6eMKcZS2BPhCAvG4n6XDIWcnUyDLurZnC3B7YKnIN+AtBXf4j9CsI992g4j4rucHJZ6amkPHxnYVge1XWi7b1So2Yw3nudcmC6c2i4hDnxSliKCj11oiXJrxvGVQUGtPmLcPrSNCCIPcpPFhYe7MbeP5QKBgQC+q9VP1hfhsDXmDZ+OL4xJl1u8PcS9bR2BpjYL7mZBJVWgF7E7mQHfdtUy8k56AvKf7W1jJGCp1U0byV8svB+MKt5vn1Q8VRfCgF5S3OOn/Zl/g6CViQXa4NTNkMi84vHFhsvzVnnLBCdt/PAxu5cOvj6ra0vTZFOGviqgzhHzUQKBgC3B1SFn525KuuvCw7KiZOPH7DcHyVI5sbL1FK3/WGFa41Nr+K+kldA2G/lFo8HHHOSmaZzWNbGRp26UjKKG936g4T2+dHCzIC34+w5lcUfYRiEZzYAGAH9n95lHwMr41cCRiChqAUvL7J12w5s2URp7SgUkSBb0WJApFQxKF4vE\"},\"certificates\":[]}")
	viper.Reset()
	viper.SetFs(ctx.Fs)
	cmd := GetMigrateCmd(ctx)
	cmd.SetArgs([]string{
		"--type", "certbot",
		"--path", acmeDir,
		"--output", outputPath,
	})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	assert.NoError(t, err)
	outputRaw, errRead := afero.ReadFile(fs, outputPath)
	assert.NoError(t, errRead)
	assert.Equal(t, string(want), string(outputRaw))
}

func TestGetMigrateRunFn_FailedWithMissingType(t *testing.T) {
	acmePath := "/app/letsencrypt/acme.json"
	outputPath := "/app/state.json"
	ctx := context.TestContext(nil)

	viper.Reset()
	viper.SetFs(ctx.Fs)
	cmd := GetMigrateCmd(ctx)
	cmd.SetArgs([]string{
		"--path", acmePath,
		"--output", outputPath,
	})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "type migration is required")
}

func TestGetMigrateRunFn_FailedWithMissingPath(t *testing.T) {
	outputPath := "/app/state.json"
	ctx := context.TestContext(nil)

	viper.Reset()
	viper.SetFs(ctx.Fs)
	cmd := GetMigrateCmd(ctx)
	cmd.SetArgs([]string{
		"--type", "traefik",
		"--output", outputPath,
	})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "source path is required")
}

func TestGetMigrateRunFn_FailedWithFileNotFound(t *testing.T) {
	acmePath := "/app/letsencrypt/acme.json"
	outputPath := "/app/state.json"
	ctx := context.TestContext(nil)
	fs := afero.NewMemMapFs()
	ctx.Fs = fs
	viper.Reset()
	viper.SetFs(ctx.Fs)
	cmd := GetMigrateCmd(ctx)
	cmd.SetArgs([]string{
		"--type", "traefik",
		"--path", acmePath,
		"--output", outputPath,
	})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read Traefik ACME file: open /app/letsencrypt/acme.json")
}
