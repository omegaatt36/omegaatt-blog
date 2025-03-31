---
title: Golang Composition over Inheritance
date: 2025-01-11
categories:
  - develop
tags:
  - golang
cover:
  image: "images/cover.png"
---

Golang 是一門簡潔有力的程式語言，相較於其他程式語言，更傾向於使用組合（composition）而不是繼承（inheritance），語言設計之初更是沒有提供繼承的關鍵字，這種設計哲學讓 Golang 在現代軟體開發中脫穎而出。

繼承固然有其優點，但在建構複雜的物件關係時，容易產生過於龐大的繼承層級結構。這使得程式碼難以閱讀和維護，就像是一棵盤根錯節的大樹，牽一髮而動全身。

過深的繼承層級會導致知名的[「脆弱基類問題」(fragile base class problem)](https://github.com/Dobiasd/articles/blob/master/implementation_inheritance_is_bad_-_the_fragile_base_class_problem.md)，使得程式碼難以修改和擴展。

組合則不同，它鼓勵建立小型、專注的 struct，然後像樂高積木一樣，將這些 struct 組合成更大的結構。這種方式讓程式碼模組化，更容易理解和修改。

## 彈性

Golang 的 type system 支援我們靈活地組合各種 struct。可以建立一個新的 struct，並在其中「嵌入」其他 struct 作為其欄位。

```go
type Car struct {
    make  string
    model string
    year  int
}

type Driver struct {
    name string
    car  Car
}

func main() {
    myCar := Car{"Toyota", "Camry", 2020}
    driver := Driver{"John", myCar}
    fmt.Println(driver.name)         // 輸出: John
    fmt.Println(driver.car.make)    // 輸出: Toyota
}
```

在這個例子中，`Driver` 透過組合 `Car` 來建立更豐富的資料結構。`Driver` 「has-a」 `Car`，而不是 「is-a」 `Car`，這提供了更高的彈性，讓 `Driver` 可以更專注在自身的邏輯。

## 程式碼複用

組合促進了程式碼的複用。可以建立許多小型、可複用的 struct，然後將它們組合成各種不同的結構。

繼承也能實現程式碼複用，但它也可能導致不必要的耦合，使得程式碼難以修改，因為對父類別的修改可能會影響到所有子類別。

```go
type Engine struct {
    power      int
    fuelType   string
}

type Wheels struct {
    count      int
    material   string
}

type Vehicle struct {
    engine Engine
    wheels Wheels
    brand  string
}

func (v Vehicle) getBrand() string {
    return v.brand
}

func (v Vehicle) getEnginePower() int {
    return v.engine.power
}
```

在這個例子中，`Vehicle` 透過組合 `Engine` 和 `Wheels` 來複用這兩個 struct 的欄位和功能。`Vehicle` 「has-a」 `Engine` and 「has-a」 `Wheels`，並可以新增自己的欄位和方法，例如 `brand`、`getBrand()` 和 `getEnginePower()`。這種方式讓 `Vehicle` 可以專注於自身的邏輯，同時又能複用 `Engine` 和 `Wheels` 的功能。

## 隱式介面

原文為：[Interfaces are implemented implicitly](https://go.dev/tour/methods/10)

除了 struct 的組合，Golang 透過**隱式介面**進一步強化了組合的優勢。不同於 Java、C# 等語言需要明確宣告實作了哪個介面，Golang 的介面是隱式實作的。

只要一個 struct 擁有了介面定義的所有方法，它就被視為實作了該介面。反之，只要沒有完全實作所有方法，就不會被視作該介面。

也可以理解成 Duck Typing：「如果它走起來像鴨子，叫起來像鴨子，那麼它就是鴨子」

```go
type Geometry interface {
    Area() float64
    Perimeter() float64
}

type Rectangle struct {
    width, height float64
}

type Circle struct {
    radius float64
}

func (r Rectangle) Area() float64 {
    return r.width * r.height
}

func (r Rectangle) Perimeter() float64 {
    return 2*r.width + 2*r.height
}

func (c Circle) Area() float64 {
    return math.Pi * c.radius * c.radius
}

func (c Circle) Perimeter() float64 {
    return 2 * math.Pi * c.radius
}

func Measure(g Geometry) {
    fmt.Println(g)
    fmt.Println(g.Area())
    fmt.Println(g.Perimeter())
}
```

在這個例子中，`Rectangle` 和 `Circle` 都沒有明確宣告自己實作了 `Geometry` 介面，但因為它們都定義了 `Area()` 和 `Perimeter()` 方法，所以它們都被視為 `Geometry`。這讓程式碼更加靈活，避免了不必要的耦合。

## 組合 vs. 繼承

「組合優於繼承」是一條廣為人知的程式設計原則：

### 耦合性：組合更鬆散

繼承是一種緊耦合的關係。子類別與父類別緊密相連，父類別的任何變動都可能影響到子類別。以下是一個 Java 的例子：

```java
public class ClassA {

    public void foo() {
    }
}

class ClassB extends ClassA {
    public void bar() {

    }
}
```

在這個例子中，`ClassB` 繼承了 `ClassA`。現在，假設 `ClassA` 的實作發生了變更，例如新增了一個 `bar()` 方法：

```java
public class ClassA {

    public void foo() {
    }

    public int bar() {
        return 0;
    }
}
```

這個變更會導致 `ClassB` 無法通過編譯，因為 `ClassB` 中已經存在一個 `bar()` 方法，但回傳型別與 `ClassA` 中的 `bar()` 方法不同。為了解決這個問題，必須修改 `ClassA` 或 `ClassB` 的程式碼。這就是繼承的緊耦合性帶來的問題，也是經典的「脆弱基類問題」。

如果使用組合，則可以避免這個問題。例如：

```java
class ClassB {
    ClassA classA = new ClassA();

    public void bar() {
        classA.foo();
        classA.bar();
    }
}
```

在這個例子中，`ClassB` 組合了 `ClassA`。即使 `ClassA` 的 `bar()` 方法發生變更，`ClassB` 也不會受到影響，因為 `ClassB` 並沒有直接繼承 `ClassA` 的 `bar()` 方法。

### 存取控制

繼承沒有提供對父類別成員的存取控制機制。子類別可以存取父類別的所有 public 和 protected 成員。這可能會導致安全問題，因為子類別可能會意外地修改父類別的狀態。組合則可以限制對內部物件的存取，提供更好的安全性。

例如，在 `ClassB` 的組合實作中，可以選擇只暴露 `ClassA` 的 `foo()` 方法：

```java
class ClassB {

    ClassA classA = new ClassA();

    public void foo() {
        classA.foo();
    }

    public void bar() {
    }

}
```

這樣，其他類別就只能透過 `ClassB` 的 `foo()` 方法來存取 `ClassA` 的 `foo()` 方法，而無法直接存取 `ClassA` 的其他成員。

## Dependency Injection

在 Golang 中，通常透過 struct 的欄位來實現組合，進而實現依賴注入。

以一個需要訪問資料庫的 `UserService` 為例：

```go
// 定義資料庫介面
type Database interface {
    GetUser(id int) (*User, error)
    SaveUser(user *User) error
}

// 定義 User struct
type User struct {
    ID   int
    Name string
}

// 定義 UserService，並透過組合注入 Database 依賴
type UserService struct {
    db Database
}

// UserService 的方法，使用注入的 db 來訪問資料庫
func (s *UserService) GetUserByID(id int) (*User, error) {
    return s.db.GetUser(id)
}

func (s *UserService) CreateUser(user *User) error {
    return s.db.SaveUser(user)
}
```

在這個例子中，`UserService` 並不關心 `Database` 具體是如何實作的，它只依賴於 `Database` 介面。

可以輕鬆地替換不同的資料庫實作，例如：

```go
// 一個 MySQL 的 Database 實作
type MySQLDatabase struct {
    // ... MySQL 連線相關的欄位
}

func (db *MySQLDatabase) GetUser(id int) (*User, error) {
    // ... 從 MySQL 資料庫中獲取使用者的程式碼
    return nil, nil
}

func (db *MySQLDatabase) SaveUser(user *User) error {
    // ... 將使用者儲存到 MySQL 資料庫的程式碼
    return nil
}

// 一個 Mock 的 Database 實作，用於測試
type MockDatabase struct{}

func (db *MockDatabase) GetUser(id int) (*User, error) {
    // ... 返回模擬的使用者資料
    return &User{ID: id, Name: "Mock User"}, nil
}

func (db *MockDatabase) SaveUser(user *User) error {
    // ... 模擬儲存使用者資料
    return nil
}
```

在實際使用時，可以根據需要注入不同的 `Database` 實作：

```go
func main() {
    // 使用 MySQLDatabase
    mysqlDB := &MySQLDatabase{}
    userService := UserService{db: mysqlDB}
    user, _ := userService.GetUserByID(1)
    fmt.Println(user)

    // 使用 MockDatabase 進行測試
    mockDB := &MockDatabase{}
    testService := UserService{db: mockDB}
    testUser, _ := testService.GetUserByID(2)
    fmt.Println(testUser)
}
```

透過組合和介面，實現了依賴注入。`UserService` 不再依賴於具體的資料庫實作，而是依賴於 `Database` 介面。

可以在不修改 `UserService` 程式碼的情況下，輕鬆地更換資料庫實作，更方便地進行單元測試。

測試時可以使用 [`gomock`](https://github.com/uber-go/mock) 庫來幫助我們快速從介面中產生 mock 結構。

## References

[golang and composition over inheritance](https://aran.dev/posts/go-and-composition-over-inheritance/)
[Dependency Injection, Duck Typing, and Clean Code in Go](https://txt.fliglio.com/2015/04/di-duck-typing-and-clean-code-in-go/)
