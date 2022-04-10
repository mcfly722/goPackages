# goPackages

## context
Replacement for standard context library. Allows to control close parent context only after child contexts.

## plugins
Module checks directory for file changes and applies this changes to engine.

## logger
Simple logger library with circular buffer. It stores events and do not block execution during logging and writing logs to storage.

## scheduler
Module allows to register many timers in one time sorted list. In main loop you just need to check nearest timer for outdating. All others are going after it.
