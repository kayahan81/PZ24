@echo off
echo Generating load...

for /l %%i in (1,1,50) do (
    curl -s http://localhost:8082/v1/tasks -H "Authorization: Bearer demo-token" > nul
    echo OK: %%i
)

for /l %%i in (1,1,20) do (
    curl -s http://localhost:8082/v1/tasks -H "Authorization: Bearer wrong-token" > nul
    echo ERR: %%i
)

echo Done!