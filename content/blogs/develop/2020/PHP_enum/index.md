---
title: PHP ENUM 偽實現
date: 2020-01-15
categories:
 - develop
tags:
 - php
aliases:
 - "/blogs/develop/2020/PHP_enum.html"
---

`Enumerations` 又可簡稱為 `Enum` ，在眾多語言中都可以讓程式碼更高效簡潔，例如我們可以在 python 中這麼宣告，並且用 `requests` 模組打一個 request，即可透過 `Enum` 來模組化 Response

``` python
import enum
import request
class HTTPResponsecode(enum.Enum)
    OK = 200
    BAD_REQUEST = 400
    NOT_FOUND = 404

r = requests.get('https://omegaatt.com/')
print(r.status_code == HTTPResponsecode.OK)
# True
```

但是在 PHP 中原生是沒有內建 `Enum` 的，必須安裝 `perl` 的 [ `SplEnum` ](https://stackoverflow.com/questions/57885011/error-class-splenum-not-found-in-php-7) 類套件庫，方能直接使用 `Enum` 的功能。或是可以透過以下的方式達到接近的效果。

## 極簡易版 Enum

首先，建立一個 `Enum` 類，一樣舉例為 HTTP 的 response code

``` php
abstract class HTTPResponsecodeEnum
{
    const OK = 200;
    const BAD_REQUEST = 400;
    const NOT_FOUND = 404;
}
```

這樣就完成最基本的 `Enum` 拉，只要透過簡單的 `if` 判斷便能輕鬆享受打包的方便

``` PHP
if($status_code == HTTPResponsecodeEnum::OK){
    // foo();
}
```

## 具有驗證功能的 Enum

根據防禦性程式寫法，我們知道千萬不要相信使用者傳過來的東西，必須去驗證資料是否正確時，極簡版本已經無法勝任。這時候可以使用 [ReflectionClass](https://www.php.net/manual/en/class.reflectionclass.php) 這個這個類別協助，由這個類來取得 Class 中的常數。

``` PHP
$oClass = new \ReflectionClass(__CLASS__);
$constants = $oClass->getConstants();
```

如此一來便可以得到該 Class 中的常數清單，試著組合一下。此時的 `constants` 使用 lazy loading，不像 `Singleton` 需要注意多線程故可以直接使用。

``` PHP
abstract class HTTPResponsecodeEnum
{
    const OK = 200;
    const BAD_REQUEST = 400;
    const NOT_FOUND = 404;

    public static function getConst(): array
    {
        static $constants;
        if (!isset($constants)) {
            $oClass = new \ReflectionClass(__CLASS__);
            $constants = $oClass->getConstants();
        }
        return $constants;
    }
}

$code = 9527;
if (in_array($code, HTTPResponsecodeEnum::getConst())) {
    // bar();
}
```

## 封裝成抽象類 Enum

但若是每個 `Enum Class` 都寫一個 `getConst` 方法似乎是太累贅了，PHP 好說歹說也是個可以寫的「很 OOP」的語言，於是乎可以把 `getConst` 與一些 function 進行整合。

``` PHP
abstract class BasicEnum {
    private static function getConst() {
        static $constantsArray;
        if (!isset($constants)) {
            self::$constantsArray = [];
        }
        $oClass = get_called_class();
        if (!array_key_exists($oClass, self::$constantsArray)) {
            $reflect = new ReflectionClass($oClass);
            self::$constantsArray[$oClass] = $reflect->getConstants();
        }
        return self::$constantsArray[$oClass];
    }

    public static function isValidName($name, $strict = false) {
        $constants = self::getConst();

        if ($strict) {
            return array_key_exists($name, $constants);
        }

        $keys = array_map('strtolower', array_keys($constants));
        return in_array(strtolower($name), $keys);
    }

    public static function isValidValue($value, $strict = true) {
        $values = array_values(self::getConst());
        return in_array($value, $values, $strict);
    }
}
```

以上改動主要是參照 [stackoverflow 上的這篇](https://stackoverflow.com/questions/254514/php-and-enumerations/21536800#21536800)，然而針對 `getConst` 是否要公開或私有這個取決於需要，透過 `isValidName` 與 `isValidValue` 已經能夠辦到了就不用再自造輪子了。未來有需要用到 `Enum` 類時，僅需繼承這個類，並如極簡版般僅需宣告常數便可以正常使用。

``` PHP
abstract class HTTPResponsecodeEnum extends BasicEnum
{
    const OK = 200;
    const BAD_REQUEST = 400;
    const NOT_FOUND = 404;
}

HTTPResponsecodeEnum::isValidName('OK');    // true
HTTPResponsecodeEnum::isValidName('NO_OK'); // false
HTTPResponsecodeEnum::isValidValue(404);    // true 
HTTPResponsecodeEnum::isValidValue(403);    //false
```

## 參考資料

* [get_called_class](https://stackoverflow.com/questions/506705/how-can-i-get-the-classname-from-a-static-call-in-an-extended-php-class)
* [寫出健壯的 PHP 應用程式(1): 防禦型程式寫法](http://asika.windspeaker.co/post/3502-strong-php-1-defensive-programming)
* [PHP and Enumerations](https://stackoverflow.com/questions/254514/php-and-enumerations/21536800#21536800)
