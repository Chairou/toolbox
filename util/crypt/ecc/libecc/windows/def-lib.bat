@echo off
REM create .def and .lib files for VS
REM https://stackoverflow.com/a/9946390

copy ..\libecc.h libecc.h
copy ..\libecc libecc.dll
dumpbin /exports libecc.dll > exports.txt
echo LIBRARY LIBECC > libecc.def
echo EXPORTS >> libecc.def
for /f "skip=19 tokens=4" %%A in (exports.txt) do @echo %%A >> libecc.def
del exports.txt

REM The librarian can use this DEF file to generate the LIB:

lib /machine:x64 /def:libecc.def /out:libecc.lib

echo Compile with the following:
echo gcc -o driver.exe driver.c -I. -L. -llibecc

gcc -o driver.exe driver.c -I. -L. -llibecc
driver.exe


REM on linux try:$ LD_LIBRARY_PATH=. ./a.out
