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

def test_wrong_location():
    with patch("business_probe.http.client.HTTPConnection") as factory:
        connection = factory.return_value
        response = connection.getresponse.return_value
        response.status = 302
        response.getheader.return_value = "http://wrong-location.com"
        
        try:
            business_probe.check_redirect(
                "test-code",
                "http://example.com",
                "http://localhost:8080",
                2
            )
        except SystemExit as e:
            assert e.code == 1
        else:
            raise AssertionError("重定向到错误的地址后，脚本没有退出")
        connection.close.assert_called_once()

def test_correct_location():
    with patch("business_probe.http.client.HTTPConnection") as factory:
        connection = factory.return_value
        response = connection.getresponse.return_value
        response.status = 302
        response.getheader.return_value = "http://example.com"
        business_probe.check_redirect(
            "test-code",
            "http://example.com",
            "http://localhost:8080",
            2
        )
        connection.close.assert_called_once()
# 正常场景不捕获 SystemExit；若意外退出，测试直接失败
test_redirect_failure(ConnectionRefusedError("模拟连接被拒绝"))
test_redirect_failure(TimeoutError("模拟连接超时"))
test_wrong_location()
test_correct_location()
print("All redirect tests passed")