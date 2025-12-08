def normalize_ticker(ticker):
    """
    Handle NSE symbols if needed.
    """
    if not ticker.endswith(".NS"):
        return ticker + ".NS"
    return ticker
