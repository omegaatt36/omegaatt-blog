---
title: Evaluate mathematical expression in PHP
date: 2020-01-16
categories:
  - develop
tags:
  - php
  - codewars
---

四則運算，不多，不就是加減乘除而已嗎，怎麼會在 [Codewars 中被標記為 2 kyu 的題目](https://www.codewars.com/kata/52a78825cdfc2cfc87000005)呢?

![](https://i.ytimg.com/vi/KTpd-CEJahw/maxresdefault.jpg)

總覺得這篇標題應該改為「Codewars 從入坑道棄坑，只需要一題就搞定」。首先我是在 2019 年底才接觸 Codewars，這種打怪練功的感覺讓我一下就上癮了，每天都在想今天能不能再拚一階，或是這題用另外一個語言會不會比較省時間。直到遇到了這題，足足卡了一整天。下面就來介紹一下四則運算在 PHP 中的解法吧。

# 先乘除後加減，看到括號先做

先舉一個簡單的例子 a+b 該怎麼做呢，有學過一點資料結構的應該很快就會回想起「中序轉後序」。

首先會需要將引入的字串轉為中序

```php
// 123 + 456 * -789
// =>
// ['123', '+', '456, '*', '-', '789']

function isOp(string $s): bool
{
    return strpos("+-*/", $s) !== false ? true : false;
}

function changeStringToInfix(string $expression): array
{
    $infix = [];
    $count = strlen($expression);
    for ($i = 0; $i < $count; $i++) {
        if (isOp($expression[$i]) === false) {
            $j = $i;
            while ($i < $count && (isOp($expression[$i]) === false)) {
                $i = $i + 1;
            }
            $infix[] = substr($expression, $j, $i - $j);
            $i = $i - 1;
        } else {
            $infix[] = substr($expression, $i, 1);
        }
    }
    return $infix;
}
```

這時候可能有眼尖的會發現，不對阿中序怎麼可以兩個運算子('\*', '-')連在一起，恭喜你踩進四則運算第一個坑：負號可能是負號也可能是減號。於是可以使用一個偷吃步的方式，將運算子後的負號一率改成另一個代表負號的特殊符號，比如「@」。而這步可以使用正規表達式輕鬆辦到。

```php
function normalization(string $expression): string
{
    // 1+-1 => 1+@1
    $expression = preg_replace("/([\+*\-*\**\/*]{1})-{1}([\d]+)/", '${1}@${2}', $expression);
    // 1+@1 => 1-1
    $expression = preg_replace("/([\+*\-*\**\/*]{1})\+@{1}([\d]+)/", '${1}-${2}', $expression);
    // 1-@1 => 1+1
    $expression = preg_replace("/([\+*\-*\**\/*]{1})\-@{1}([\d]+)/", '${1}${2}', $expression);
    return $expression;
}
```

此時再回去看剛剛的例子，先進行正規化後就可以得到正常的中序式了。

# 中序轉後序

網路上一堆中序轉後序的教學這邊就不再贅述了，稍微提供一下我的方法。

```php
function getOpSize(string $s): int
{
    switch ($s) {
        case "+":
            return 1;
        case "-":
            return 1;
        case "*":
            return 2;
        case "/":
            return 2;
        default:
            return -1;
    }
}

function changeInfixtToPostfix(array $infix): array
{
    $stack = [];
    $postfix = [];
    foreach ($infix as $s) {
        if (isOp($s)) {
            if (count($stack) == 0) {
                $stack[] = $s;
            } else {
                if (getOpSize($s) > getOpSize(end($stack))) {
                    $stack[] = $s;
                } else {
                    $postfix[] = array_pop($stack);
                    $stack[] = $s;
                }
            }
        } else {
            $postfix[] = $s;
        }
    }
    while (count($stack) > 0) {
        $postfix[] = array_pop($stack);
    }
    return $postfix;
}
```

# 處理後序式

後序式的處理就比較簡單了，只要注意有把負號替換成「@」所以在計算上需要多一道工。

```PHP
function execPostfix(array $postfix)
{
    $stack = [];
    foreach ($postfix as $s) {
        if (!isOp($s)) {
            $stack[] = $s;
        } else {
            $b = array_pop($stack);
            $b = strpos($b, '@') !== false ? substr($b, 1) * -1 : $b;
            $a = array_pop($stack);
            $a = strpos($a, '@') !== false ? substr($a, 1) * -1 : $a;
            switch ($s) {
                case "+":
                    $r = $a + $b;
                    break;
                case "-":
                    $r = $a - $b;
                    break;
                case "*":
                    $r = $a * $b;
                    break;
                case "/":
                    $r = $a / $b;
                    break;
            }
            // echo $a . ' ' . $s . ' ' . $b .' = ' . $r . PHP_EOL;
            $stack[] = $r;
        }
    }
    $value = end($stack);
    return strpos($value, '@') !== false ? substr($value, 1) * -1 : $value;;
}
```

# 好像忘了什麼

如果只是到這裡那這題可能只有個 4 kyu 甚至 5 kyu，四則運算中還有個括號呢，這裡提供一個思考的方向

```PHP
1-(-2) => 1--2 => 1+2 => 3
```

有沒有發現其實只要把括號內的部分先計算完再拆開就可以進行剛剛的計算流程了，然而也可能遇到括弧內有計算式的(正常應該都會有吧)

```PHP
1-(-(-1*-3))
```

一開始我看到這種 test case 真的是氣到說不出話來，但一個一個括號慢慢拆開來，一行式子遲早會被做完。

```PHP
while (strpos($expression, '(') !== false) {
    $expression = removeParentheses($expression);
}

function removeParentheses(string $expression): string
{
    $left_index = strrpos($expression, '(');
    $right_index = strpos(substr($expression, $left_index), ')') + $left_index;
    $left = substr($expression, 0, $left_index);
    $right = substr($expression, $right_index + 1);
    $sub_expression = substr($expression, $left_index + 1, $right_index - $left_index - 1);
    $sub_expression =  normalization($sub_expression);
    $sub_infix = changeStringToInfix($sub_expression);
    $sub_postfig = changeInfixtToPostfix($sub_infix);
    $sub_value = execPostfix($sub_postfig);
    return $left . $sub_value . $right;
}
```

# 最終成果

最終只要將上述的流程拚在一起，這題就解出來啦~~~

```PHP
function calc(string $expression): float
{
    $expression = str_replace(' ', '', $expression);
    $expression = str_replace('--', '+', $expression);
    while (strpos($expression, '(') !== false) {
        $expression = removeParentheses($expression);
    }

    $expression =  normalization($expression);
    $infix = changeStringToInfix($expression);
    $postfix = changeInfixtToPostfix($infix);
    $value = execPostfix($postfix);
    return sprintf("%.6f", $value);
}
```

解完後原本很暢快的心情，在看到大神只靠正規表達式用二十行解完後頓時又失去信心了(不過那個誰看的懂拉)，如果覺得我的 code 哪個部分不清楚或可以更好的歡迎留言指教。
