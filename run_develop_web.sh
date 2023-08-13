source develop.env

function cleanup() {
    rm -f my-life-in-weeks-web
}
trap cleanup EXIT


# Compile Go
GO111MODULE=on GOGC=off go build -mod=vendor -v -o my-life-in-weeks-web ./cmd/web/
./my-life-in-weeks-web