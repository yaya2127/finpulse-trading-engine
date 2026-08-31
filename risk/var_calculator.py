import math
import random

class RiskEngine:
    """
    Real-Time Portfolio Value-at-Risk (VaR), Monte Carlo Simulation, 
    and Black-Scholes Options Pricing & Greeks Engine.
    """
    
    @staticmethod
    def calculate_var_95(portfolio_value, volatility_daily, confidence_z=1.645):
        """Calculates 95% Confidence Value at Risk (VaR)"""
        var_amount = portfolio_value * volatility_daily * confidence_z
        return round(var_amount, 2)
    
    @staticmethod
    def calculate_sharpe_ratio(returns, risk_free_rate=0.04):
        """Calculates Annualized Sharpe Ratio"""
        if not returns or len(returns) < 2:
            return 0.0
        avg_return = sum(returns) / len(returns)
        variance = sum((r - avg_return) ** 2 for r in returns) / (len(returns) - 1)
        std_dev = math.sqrt(variance) if variance > 0 else 0.0001
        
        annualized_return = avg_return * 252
        annualized_std = std_dev * math.sqrt(252)
        
        return round((annualized_return - risk_free_rate) / annualized_std, 2)
    
    @staticmethod
    def run_monte_carlo(current_price, days=30, num_paths=1000, mu=0.08, sigma=0.25):
        """Runs 1,000 Monte Carlo price trajectory simulations using Geometric Brownian Motion"""
        dt = 1.0 / 252.0
        paths = []
        
        for _ in range(num_paths):
            price = current_price
            trajectory = [price]
            for _ in range(days):
                epsilon = random.gauss(0, 1)
                drift = (mu - 0.5 * (sigma ** 2)) * dt
                shock = sigma * math.sqrt(dt) * epsilon
                price *= math.exp(drift + shock)
                trajectory.append(round(price, 2))
            paths.append(trajectory)
            
        final_prices = [p[-1] for p in paths]
        var_95_simulated = round(current_price - sorted(final_prices)[int(num_paths * 0.05)], 2)
        
        return {
            'initial_price': current_price,
            'simulated_var_95': var_95_simulated,
            'expected_price_30d': round(sum(final_prices) / num_paths, 2),
            'max_simulated_price': max(final_prices),
            'min_simulated_price': min(final_prices)
        }

class BlackScholesGreeks:
    """
    Black-Scholes Options Pricing & Greeks Calculator (Delta, Gamma, Vega, Theta)
    Optimized with Pre-Computed Cumulative Normal Distribution Lookup Table (LUT) for sub-microsecond latency.
    """
    # Pre-computed 64-bit IEEE floating-point Lookup Table for Normal CDF [-4.0 to 4.0]
    _LUT_MIN = -4.0
    _LUT_MAX = 4.0
    _LUT_STEPS = 1000
    _LUT_STEP_SIZE = (_LUT_MAX - _LUT_MIN) / _LUT_STEPS
    _CDF_LUT = [(1.0 + math.erf((_LUT_MIN + i * _LUT_STEP_SIZE) / math.sqrt(2.0))) / 2.0 for i in range(_LUT_STEPS + 1)]

    @classmethod
    def norm_cdf(cls, x):
        """Fast lookup-table accelerated normal CDF calculation"""
        if x <= cls._LUT_MIN: return 0.0
        if x >= cls._LUT_MAX: return 1.0
        idx = int((x - cls._LUT_MIN) / cls._LUT_STEP_SIZE)
        return cls._CDF_LUT[idx]

    @staticmethod
    def norm_pdf(x):
        return math.exp(-0.5 * x * x) / math.sqrt(2.0 * math.pi)

    @classmethod
    def calculate_option(cls, S, K, T, r, sigma, option_type="CALL"):
        """
        S: Spot Price, K: Strike Price, T: Time to Expiry (years), 
        r: Risk-free rate, sigma: Volatility
        """
        if T <= 0 or sigma <= 0:
            return {"price": max(0, S - K) if option_type == "CALL" else max(0, K - S), "delta": 0, "gamma": 0, "vega": 0, "theta": 0}
            
        d1 = (math.log(S / K) + (r + 0.5 * sigma ** 2) * T) / (sigma * math.sqrt(T))
        d2 = d1 - sigma * math.sqrt(T)

        if option_type == "CALL":
            price = S * cls.norm_cdf(d1) - K * math.exp(-r * T) * cls.norm_cdf(d2)
            delta = cls.norm_cdf(d1)
        else:
            price = K * math.exp(-r * T) * cls.norm_cdf(-d2) - S * cls.norm_cdf(-d1)
            delta = cls.norm_cdf(d1) - 1.0

        gamma = cls.norm_pdf(d1) / (S * sigma * math.sqrt(T))
        vega = S * cls.norm_pdf(d1) * math.sqrt(T) / 100.0
        theta = (- (S * cls.norm_pdf(d1) * sigma) / (2.0 * math.sqrt(T)) - r * K * math.exp(-r * T) * cls.norm_cdf(d2)) / 365.0

        return {
            "spot": S, "strike": K, "expiry_years": T,
            "option_type": option_type,
            "price": round(price, 4),
            "delta": round(delta, 4),
            "gamma": round(gamma, 4),
            "vega": round(vega, 4),
            "theta": round(theta, 4)
        }

if __name__ == "__main__":
    risk = RiskEngine.run_monte_carlo(100.0)
    greeks = BlackScholesGreeks.calculate_option(100.0, 105.0, 0.25, 0.05, 0.20, "CALL")
    print("Monte Carlo Result:", risk)
    print("Black-Scholes Call Greeks:", greeks)
