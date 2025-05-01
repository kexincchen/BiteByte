const ProductCard = ({ product, availability }) => {
  return (
    <div className="product-card">
      <h3>{product.name}</h3>
      <p>${product.price.toFixed(2)}</p>
      
      {availability === false && (
        <span className="sold-out-badge">Sold Out</span>
      )}
      
      <button 
        onClick={() => addToCart(product)}
        disabled={availability === false}
        className="add-to-cart-button"
      >
        {availability === false ? "Sold Out" : "Add to Cart"}
      </button>
    </div>
  );
}; 