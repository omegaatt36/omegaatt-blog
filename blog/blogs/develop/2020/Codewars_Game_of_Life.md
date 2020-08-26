---
title: 威康生命遊戲 in PHP
date: 2020-01-23
tags:
  - php
  - codewars
categories:
 - develop
---

# 威康生命遊戲 in PHP

[![生命遊戲：另一種計算機 |混亂博物館](https://i.ytimg.com/vi/GQNREcMVPHY/hqdefault.jpg)](https://www.youtube.com/embed/GQNREcMVPHY "生命遊戲：另一種計算機 |混亂博物館")

生命遊戲，無論在程式、哲學、生命探討上都十分有趣，在生命遊戲中甚至出現永動機。而在 Codewars 中也有生命遊戲的 [題目](https://www.codewars.com/kata/52423db9add6f6fc39000354)，規則也十分簡單，以下為維基百科的內容：

```
1. 每個細胞有兩種狀態 - 存活或死亡，每個細胞與以自身為中心的周圍八格細胞產生互動（如圖，黑色為存活，白色為死亡）
2. 當前細胞為存活狀態時，當周圍的存活細胞低於2個時（不包含2個），該細胞變成死亡狀態。（模擬生命數量稀少）
3. 當前細胞為存活狀態時，當周圍有2個或3個存活細胞時，該細胞保持原樣。
4. 當前細胞為存活狀態時，當周圍有超過3個存活細胞時，該細胞變成死亡狀態。（模擬生命數量過多）
5. 當前細胞為死亡狀態時，當周圍有3個存活細胞時，該細胞變成存活狀態。（模擬繁殖）
```

若想知道關於更多可以觀看上方由**混亂博物館**出品的影片，此篇就針對這 4kyu 的題目進行探討解法。首先題目中指出這個宇宙的 x,y 軸都是無限大的，所以生成高斯帕機槍且迭代過多次可能造成記憶體使用過量。

# 迭代

首先題目給定一個二維矩陣以及迭代次數(世代)，而需取得過了該迭代次數後的情形。

```PHP
function get_generation(array $cells, int $generations): array
{
    if($cells == [])
    return [[]];
    $w = count($cells);
    $h = count($cells[0])
    for ($index = 0; $index < $generations; $index++) {
        $cells = getNextGen($cells, $w, $h);
        $sum = 0;
        foreach ($cells as $row) {
            $sum += array_sum($row);
        }
        if ($sum == 0) {
            return [[]];
        }
    }

    return $cells;
}
```

於是可以建立一個方法 getNextGen 來獲取下個世代的結果

```PHP
function getNextGen(array $cells, $height, $width): array
{
    $cells = appendCell($cells, $height, $width);
    $height = $height+2;
    $width = $width +2;
    $kill_queue = $born_queue = [];
    for ($y = 0; $y < $height; $y++) {
        for ($x = 0; $x < $width; $x++) {
            $neighbor_count = getAliveNeighborCount($x, $y, $cells, $height, $width);
            // rule 2,4
            if ($cells[$y][$x] && ($neighbor_count < 2 || $neighbor_count > 3)) {
                $kill_queue[] = [$y, $x];
            }
            // rule 5
            if (!$cells[$y][$x] && $neighbor_count === 3) {
                $born_queue[] = [$y, $x];
            }
        }
    }
    // rule 2,4
    foreach ($kill_queue as $c) {
        $cells[$c[0]][$c[1]] = 0;
    }
    // rule 5
    foreach ($born_queue as $c) {
        $cells[$c[0]][$c[1]] = 1;
    }
    // remove empty margin
    $cells = shiftCell($cells, $height, $width);

    return $cells;
}
```

但由於此版本是無限版 ( Unlimited Edition )，在根據規則進行演化前，需要先將世界向四方擴張一格。

```PHP
function appendCell(array $cells, $height, $width)
{
    $new_cells = array_fill(0, $height + 2, array_fill(0, $width + 2, 0));
    for ($y = 0; $y < $height; $y++) {
        for ($x = 0; $x < $width; $x++) {
            $new_cells[$y + 1][$x + 1] = $cells[$y][$x];
        }
    }
    return $new_cells;
}
```

# 取得鄰近細胞數量

這邊塞入整個 $cells 而不是用先透過 array_slice 來取得 9x9 就是使用空間換時間大法(X)，若是實際情況就需要看針對迭代次數、二維大小以及機器性能進行優化了。

```PHP
function getAliveNeighborCount($x, $y, $cells, $height, $width)
{
    $alive_count = 0;
    for ($y2 = $y - 1; $y2 <= $y + 1; $y2++) {
        if ($y2 < 0 || $y2 >= $height) {
            continue;
        }
        for ($x2 = $x - 1; $x2 <= $x + 1; $x2++) {
            if ($x2 == $x && $y2 == $y) {
                continue;
            }
            if ($x2 < 0 || $x2 >= $width) {
                continue;
            }
            if ($cells[$y2][$x2]) {
                $alive_count += 1;
            }
        }
    }
    return $alive_count;
}
```

# 縮小宇宙

而迭代玩一次需要進行針對邊界的縮小，建立一個方法 shiftCell 不斷回 call 確保上、下、左、右皆無空欄列。

```PHP
function shiftCell(array $cells, $height, $width)
{
    if($cells[0] == null){
        return [[]];
    }
    // top
    if (array_sum($cells[0]) == 0) {
        return shiftCell(array_slice($cells, 1), $height - 1, $width);
    }
    // bottom
    if (array_sum($cells[$height - 1]) == 0) {
        return shiftCell(array_slice($cells, 0, $height - 1), $height - 1, $width);
    }
    // left
    $left_count = 0;
    for ($y = 0; $y < $height; $y++) {
        $left_count = $left_count + $cells[$y][0];
    }
    if ($left_count == 0) {
        foreach ($cells as &$row) {
            array_shift($row);
        }
        return shiftCell($cells, $height, $width - 1);
    }
    // right
    $right_count = 0;
    for ($y = 0; $y < $height; $y++) {
        $right_count = $right_count + $cells[$y][$width - 1];
    }
    if ($right_count == 0) {
        foreach ($cells as &$row) {
            $row = array_slice($row, 0, $width - 1);
        }
        return shiftCell($cells, $height, $width - 1);
    }

    return $cells;
}
```

混亂遊戲著實是一個有趣的題目，僅有四(五)條規則就能模擬細胞演化(細胞自動機)，有會死亡的、會繁殖的，穩定以及不穩定的、永生的，那是否整個宇宙也能透過一條公式詮釋呢。