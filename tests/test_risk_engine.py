import unittest
from risk.var_calculator import RiskEngine, BlackScholesGreeks

class TestRiskEngine(unittest.TestCase):
    
    def test_var_95(self):
        var = RiskEngine.calculate_var_95(100000.0, 0.02)
        self.assertGreater(var, 0)
        self.assertEqual(var, 3290.0)

    def test_sharpe_ratio(self):
        returns = [0.01, 0.02, -0.005, 0.015, 0.03, -0.01, 0.025]
        sharpe = RiskEngine.calculate_sharpe_ratio(returns)
        self.assertIsInstance(sharpe, float)

    def test_monte_carlo(self):
        sim = RiskEngine.run_monte_carlo(150.0, days=10, num_paths=100)
        self.assertIn('simulated_var_95', sim)
        self.assertGreater(sim['max_simulated_price'], sim['min_simulated_price'])

    def test_black_scholes_call(self):
        call = BlackScholesGreeks.calculate_option(100.0, 100.0, 0.5, 0.05, 0.2, "CALL")
        self.assertGreater(call['price'], 0)
        self.assertGreater(call['delta'], 0)
        self.assertGreater(call['gamma'], 0)

if __name__ == '__main__':
    unittest.main()
