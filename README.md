# gopatch
Golang REST patch operation.

##### a simple library to do patch operation on REST.

##### Getting started
Install `gopatch` library with go get:

    go get github.com/yuseferi/gopatch

##### Sample code:

....

##### A sample :


    {
        "op": "replace",
        "path": "/field1",
        "value": "NewValue"
    }

it does `replace` operation on `field1` and change it's value to `NewValue`

##### License
Code is distributed under MIT license, feel free to use it in your proprietary projects as well.