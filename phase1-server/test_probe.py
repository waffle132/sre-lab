import probe
import sys

def test_is_healthy():
    assert probe.is_healthy(200, "ERROR") == False
    print("Test 1 passed")
    assert probe.is_healthy(200, "OK\n") == True
    print("Test 2 passed")
test_is_healthy()    
    