## glox - A Lox language interpreter written in Go

Following along with https://craftinginterpreters.com/. Not guaranteed to be good or working.

Current lox grammar:

```
program        → declaration* EOF ;

declaration    → classDecl
               | funDecl
               | varDecl
               | statement ;

classDecl      → "class" IDENTIFIER ( "<" IDENTIFIER )?
               "{" function* "}" ;

funDecl        → "fun" function ;
function       → IDENTIFIER "(" parameters? ")" block ;
parameters     → IDENTIFIER ( "," IDENTIFIER )* ;


varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;

statement      → exprStmt
               | breakStmt
               | continueStmt
               | forStmt
               | ifStmt
               | printStmt
               | whileStmt
               | whileStmt
               | block;

exprStmt       → expression ";" ;
breakStmt      → "break" ";" ;
continueStmt   → "continue" ";" ;
forStmt        → "for" "(" ( varDecl | exprStmt | ";" )
                 expression? ";"
                 expression? ")" statement ;
ifStmt         → "if" "(" expression ")" statement
               ( "else" statement )? ;
printStmt      → "print" expression ";" ;
returnStmt     → "return" expression? ";" ;
whileStmt      → "while" "(" expression ")" statement ;
block          → "{" declaration* "}" ;

expression     → assignment ;
assignment     → ( call ". " )? IDENTIFIER "=" assignment
               | condition ;
condition      → logic_or ( ( "?" ) condition ( ":" ) condition )? ;
logic_or       → logic_and ( "or" logic_and )* ;
logic_and      → equality ( "and" equality )* ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary
               | call ;
call           → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
arguments      → expression ( "," expression )* ;
primary        → NUMBER | STRING | "true" | "false" | "nil"
               | IDENTIFIER | "(" expression ")" | "super" "." IDENTIFIER ;
```
