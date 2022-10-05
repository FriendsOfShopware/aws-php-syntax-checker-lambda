# PHP Syntax Linter in AWS Lambda

This is a PHP syntax checker that runs in AWS Lambda. It is a simple wrapper around the PHP linter, `php -l`.


## Usage

```bash
# PHP 7.2
curl -X POST "https://php-syntax-checker.fos.gg/?version=7.2" -F file=@/FroshTools.zip

# PHP 7.4
curl -X POST "https://php-syntax-checker.fos.gg/?version=7.4" -F file=@/FroshTools.zip

# PHP 8.1
curl -X POST "https://php-syntax-checker.fos.gg/?version=8.1" -F file=@/FroshTools.zip
```

## Responses

### Success

HTTP Status: 200

```json
{"message": "All files are valid"}
```

### Error

HTTP Status: 409

```json
{"errors":["File FroshTools/src/Command/ChangeUserPasswordCommand.php: \nParse error: syntax error, unexpected 'EntityRepositoryInterface' (T_STRING), expecting function (T_FUNCTION) or const (T_CONST) in /tmp/php-syntax-checker1065232478 on line 20\nErrors parsing /tmp/php-syntax-checker1065232478\n"]}
```

## Compile PHP

```bash
docker run --rm -v $(pwd):/out amazonlinux:1 bash

yum install gcc g++ make tar gzip wget

#download source

./configure --disable-all --disable-cgi --enable-cli --enable-static

make -j

cp sapi/cli/php out/phpVERSION
```