source develop.env

function cleanup() {
    rm -f my-life-in-weeks
}
trap cleanup EXIT


# Compile Go
GO111MODULE=on GOGC=off go build -mod=vendor -v -o my-life-in-weeks ./cmd/import/
./my-life-in-weeks