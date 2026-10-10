import business_probe
from unittest.mock import patch

def test_redirect_failure(error):
    with patch("business_probe.http.client.HTTPConnection") as factory:
        connection = factory.return_value
        connection.request.side_effect = error
        try:
            business_probe.check_redirect("test-code", "http://example.com", "http://localhost:8080", 2)
        except SystemExit as e:
            assert e.code == 1
        else:
            raise AssertionError("连接失败后，脚本没有退出")
        connection.close.assert_called_once()
test_redirect_failure(ConnectionRefusedError("模拟连接被拒绝"))
test_redirect_failure(TimeoutError("模拟连接超时"))
print("Failure handling test passed")