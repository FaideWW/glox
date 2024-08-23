## glox - A Lox language interpreter written in Go

Following along with https://craftinginterpreters.com/

Current lox grammar:

```
program        → declaration* EOF ;

declaration    → varDecl | statement ;

varDecl        → "var" IDENTIFIER ( "=" expression )? ";" ;

statement      → exprStmt
               | forStmt
               | ifStmt
               | printStmt
               | whileStmt
               | continueStmt
               | breakStmt
               | block;

forStmt        → "for" "(" ( varDecl | exprStmt | ";" )
                 expression? ";"
                 expression? ")" statement ;

ifStmt         → "if" "(" expression ")" statement
               ( "else" statement )? ;
exprStmt       → expression ";" ;
printStmt      → "print" expression ";" ;
whileStmt      → "while" "(" expression ")" statement ;
continueStmt   → "continue" ";" ;
breakStmt      → "break" ";" ;
block          → "{" declaration* "}" ;

expression     → assignment ( ( "," ) assignment )* ;
assignment     → IDENTIFIER "=" assignment
               | condition ;
condition      → logic_or ( ( "?" ) condition ( ":" ) condition )? ;
logic_or       → logic_and ( "or" logic_and )* ;
logic_and      → equality ( "and" equality )* ;
equality       → comparison ( ( "!=" | "==" ) comparison )* ;
comparison     → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
term           → factor ( ( "-" | "+" ) factor )* ;
factor         → unary ( ( "/" | "*" ) unary )* ;
unary          → ( "!" | "-" ) unary
               | primary ;
primary        → NUMBER | STRING | "true" | "false" | "nil"
               | "(" expression ")"
               | IDENTIFIER ;
```
