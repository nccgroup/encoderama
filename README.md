# encoderama

`encoderama` is a simple tool for encoding input strings or wordlists then using the output to, primarily, fuzz web applications. 

## Building

Assuming you have golang installed...

git clone and `go build`

## Usage

To see the usuage run `./encodarama` without any arguments.

The default encoding types are:

* Output plain (p)
* URL encode key characters (u)
* HTML encode key characters (h)

## Examples

URL encode and double URL encode key characters (also plaintext):

```bash
./encoderama -f /tmp/input.txt -e p,u,du
Hello World
% is a key char
So is +
Hello World
% is a key char
So is +
Hello+World
%25+is+a+key+char
So+is+%2B
Hello%2BWorld
%2525%2Bis%2Ba%2Bkey%2Bchar
So%2Bis%2B%252B
```

URL encode an injection test string but build up the string one char at a time:

```bash
./encoderama -i -e u '">;!--\"<xss>=&{()}'
%22
%22%3E
%22%3E%3B
%22%3E%3B%21
%22%3E%3B%21-
%22%3E%3B%21--
%22%3E%3B%21--%5C
%22%3E%3B%21--%5C%22
%22%3E%3B%21--%5C%22%3C
%22%3E%3B%21--%5C%22%3Cx
%22%3E%3B%21--%5C%22%3Cxs
%22%3E%3B%21--%5C%22%3Cxss
%22%3E%3B%21--%5C%22%3Cxss%3E
%22%3E%3B%21--%5C%22%3Cxss%3E%3D
%22%3E%3B%21--%5C%22%3Cxss%3E%3D%26
%22%3E%3B%21--%5C%22%3Cxss%3E%3D%26%7B
%22%3E%3B%21--%5C%22%3Cxss%3E%3D%26%7B%28
%22%3E%3B%21--%5C%22%3Cxss%3E%3D%26%7B%28%29
%22%3E%3B%21--%5C%22%3Cxss%3E%3D%26%7B%28%29%7D
```

