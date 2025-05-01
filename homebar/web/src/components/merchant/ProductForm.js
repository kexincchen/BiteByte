import React, { useState, useEffect, useContext } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { productAPI } from "../../services/api";
import { AuthContext } from "../../contexts/AuthContext";
import { ingredientAPI } from "../../services/api";
import { productIngredientAPI } from "../../services/api";

const ProductForm = ({ isEditing = false }) => {
  const { currentUser } = useContext(AuthContext);
  const navigate = useNavigate();
  const { id } = useParams();

  const [formData, setFormData] = useState({
    name: "",
    description: "",
    price: "",
    category: "",
    file: null,
    is_available: true,
  });

  const [existingImageUrl, setExistingImageUrl] = useState('');

  const [loading, setLoading] = useState(isEditing);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");
  const [categories, setCategories] = useState([
    "Cocktails",
    "Spirits",
    "Beer",
    "Wine",
    "Non-Alcoholic",
    "Snacks",
  ]);

  // Add states for ingredients
  const [ingredients, setIngredients] = useState([]);
  const [productIngredients, setProductIngredients] = useState([]);
  const [selectedIngredient, setSelectedIngredient] = useState("");
  const [quantity, setQuantity] = useState(1);

  useEffect(() => {
    const fetchProductCategories = async () => {
      try {
        // In a real app, you would fetch categories from the server
        // For now, we'll use the hardcoded categories
      } catch (error) {
        console.error("Error fetching categories:", error);
      }
    };

    const fetchProductDetails = async () => {
      if (isEditing && id) {
        try {
          const response = await productAPI.getProduct(id);
          console.log(response.data);
          setFormData(response.data);
          setFormData({
            name:        response.data.name,
            description: response.data.description,
            price:       response.data.price,
            category:    response.data.category,
            file:        null,
            is_available:response.data.is_available
          });
          setExistingImageUrl(
            `http://localhost:8080/api/products/${id}/image?ts=${Date.now()}`
          );
          setLoading(false);
        } catch (error) {
          console.error("Error fetching product details:", error);
          setError("Failed to load product details");
          setLoading(false);
        }
      }
    };

    // Fetch all ingredients for the merchant
    const fetchIngredients = async () => {
      try {
        // Get merchant ID from currentUser
        const merchantId = currentUser.merchant_id;
        console.log("Merchant ID:", merchantId);
        const response = await ingredientAPI.getIngredients(merchantId);
        setIngredients(response.data);
      } catch (error) {
        console.error("Error fetching ingredients:", error);
      }
    };

    // Fetch product ingredients (only when editing)
    const fetchProductIngredients = async () => {
      if (isEditing && id) {
        try {
          const response = await productIngredientAPI.getProductIngredients(id);
          console.log(response.data);
          setProductIngredients(response.data || []);
        } catch (error) {
          console.error("Error fetching product ingredients:", error);
        }
      }
    };

    fetchProductCategories();
    fetchIngredients();

    if (isEditing) {
      fetchProductDetails();
      fetchProductIngredients();
    }
  }, [isEditing, id]);

  const handleDelete = async (productId) => {
    if (window.confirm("Are you sure you want to delete this product?")) {
      try {
        await productAPI.deleteProduct(productId);
        navigate("/merchant/products");
      } catch (error) {
        console.error("Error deleting product:", error);
        alert("Failed to delete product. Please try again.");
      }
    }
  };

  const handleChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value,
    }));
  };

  const handleFileChange = (e) => {
    const file = e.target.files?.[0] || null;
    setFormData((prev) => ({ ...prev, file }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();

    if (!currentUser || currentUser.role !== "merchant") {
      setError("You must be logged in as a merchant to perform this action");
      return;
    }

    setSubmitting(true);
    setError("");

    try {
      // Format the data for submission
      const multipart = new FormData();
      multipart.append("name", formData.name);
      multipart.append("description", formData.description);
      multipart.append("price", formData.price);
      multipart.append("category", formData.category);
      multipart.append("is_available", formData.is_available);
      multipart.append("merchant_id", currentUser.merchant_id);
      if (formData.file) multipart.append("image", formData.file);

      let productId;
      if (isEditing) {
        await productAPI.updateProduct(id, multipart);
        productId = id;
      } else {
        const response = await productAPI.createProduct(multipart);
        productId = response.data.id;
      }

      // Process ingredients for the product
      if (productIngredients.length > 0) {
        // For new products, or if we've modified ingredients for existing products
        for (const ingredient of productIngredients) {
          try {
            // Use addIngredientToProduct which handles both creates and updates
            await productIngredientAPI.addIngredientToProduct(productId, {
              product_id: parseInt(productId),
              ingredient_id: parseInt(ingredient.ingredient_id),
              quantity: parseFloat(ingredient.quantity),
            });
          } catch (ingredientError) {
            console.error("Error saving ingredient:", ingredientError);
            // Continue with other ingredients even if one fails
          }
        }
      }

      navigate("/merchant/products");
    } catch (error) {
      console.error("Error saving product:", error);
      setError("Failed to save product. Please try again.");
    } finally {
      setSubmitting(false);
    }
  };

  // Add ingredient to product
  const handleAddIngredient = () => {
    console.log("Adding ingredient to product: ", selectedIngredient, quantity);
    if (!selectedIngredient || quantity <= 0) return;

    const ingredient = ingredients.find(
      (i) => i.id === parseInt(selectedIngredient)
    );

    console.log("Ingredient: ", ingredient);

    if (ingredient) {
      const newIngredient = {
        ingredient_id: ingredient.id,
        ingredient_name: ingredient.name,
        ingredient_unit: ingredient.unit,
        quantity: parseFloat(quantity),
      };

      setProductIngredients([...productIngredients, newIngredient]);
      setSelectedIngredient("");
      setQuantity(1);
    }
  };

  // Remove ingredient from selection
  const handleRemoveIngredient = async (ingredientId) => {
    console.log("Removing ingredient from product: ", ingredientId);
    if (isEditing) {
      try {
        await productIngredientAPI.removeIngredientFromProduct(
          id,
          ingredientId
        );
      } catch (error) {
        console.error("Error removing ingredient:", error);
        return;
      }
    }

    setProductIngredients(
      productIngredients.filter((item) => item.ingredient_id !== ingredientId)
    );
  };

  // Update ingredient quantity
  const handleQuantityChange = (ingredientId, newQuantity) => {
    console.log("Updating ingredient quantity: ", ingredientId, newQuantity);
    setProductIngredients(
      productIngredients.map((item) =>
        item.ingredient_id === ingredientId
          ? { ...item, quantity: parseFloat(newQuantity) }
          : item
      )
    );

    // If editing, update on the server
    if (isEditing) {
      const updatedItem = productIngredients.find(
        (item) => item.ingredient_id === ingredientId
      );
      if (updatedItem) {
        productIngredientAPI
          .updateProductIngredient(
            parseInt(id),
            parseInt(ingredientId),
            parseFloat(newQuantity)
          )
          .catch((error) => {
            console.error("Error updating ingredient quantity:", error);
          });
      }
    }
  };

  // Get filtered list of ingredients that aren't already added
  const availableIngredients = ingredients.filter(
    (ingredient) =>
      !productIngredients.some((pi) => pi.ingredient_id === ingredient.id)
  );

  if (loading) return <div>Loading product details...</div>;

  return (
    <div className="product-form-container">
      <h1>{isEditing ? "Edit Product" : "Add New Product"}</h1>

      {error && <div className="error">{error}</div>}

      <form onSubmit={handleSubmit} className="product-form">
        <div className="form-group">
          <label htmlFor="name">Product Name</label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleChange}
            required
          />
        </div>

        <div className="form-group">
          <label htmlFor="description">Description</label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleChange}
            rows="4"
            required
          />
        </div>

        <div className="form-row">
          <div className="form-group">
            <label htmlFor="price">Price ($)</label>
            <input
              type="number"
              id="price"
              name="price"
              value={formData.price}
              onChange={handleChange}
              step="0.01"
              min="0"
              required
            />
          </div>

          <div className="form-group">
            <label htmlFor="category">Category</label>
            <select
              id="category"
              name="category"
              value={formData.category}
              onChange={handleChange}
              required
            >
              <option value="">Select a category</option>
              {categories.map((category) => (
                <option key={category} value={category}>
                  {category}
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="form-group">
          <label htmlFor="image">Image (jpg/png)</label>
          <input
            type="file"
            id="image"
            accept="image/png, image/jpeg"
            onChange={handleFileChange}
          />
          {(formData.file || existingImageUrl) && (
             <div className="image-preview">
               <img
                 src={
                   formData.file
                     ? URL.createObjectURL(formData.file)
                         : existingImageUrl
                     }
                 alt="preview"
                 style={{ maxWidth: 200 }}
               />
             </div>
          )}
        </div>

        <div className="form-group checkbox-group">
          <input
            type="checkbox"
            id="is_available"
            name="is_available"
            checked={formData.is_available}
            onChange={handleChange}
          />
          <label htmlFor="is_available">Available for purchase</label>
        </div>

        {/* Ingredients Section */}
        <h2>Product Ingredients</h2>
        <div className="ingredients-section">
          {/* Current Ingredients Table */}
          <div className="ingredients-table">
            <table>
              <thead>
                <tr>
                  <th>Ingredient</th>
                  <th>Quantity</th>
                  <th>Unit</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {productIngredients.length === 0 ? (
                  <tr>
                    <td colSpan="4" className="no-ingredients">
                      No ingredients added
                    </td>
                  </tr>
                ) : (
                  productIngredients.map((item) => (
                    <tr key={item.ingredient_id}>
                      <td>{item.ingredient_name}</td>
                      <td>
                        <input
                          type="number"
                          className="quantity-input"
                          value={item.quantity}
                          onChange={(e) =>
                            handleQuantityChange(
                              item.ingredient_id,
                              e.target.value
                            )
                          }
                          step="0.1"
                          min="0.1"
                        />
                      </td>
                      <td>{item.ingredient_unit}</td>
                      <td>
                        <button
                          type="button"
                          className="remove-ingredient-btn"
                          onClick={() =>
                            handleRemoveIngredient(item.ingredient_id)
                          }
                        >
                          Remove
                        </button>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>

          {/* Add Ingredients Form */}
          <div className="add-ingredient-form">
            <select
              value={selectedIngredient}
              onChange={(e) => setSelectedIngredient(e.target.value)}
              className="ingredient-select"
            >
              <option value="">Select an ingredient</option>
              {availableIngredients.map((ingredient) => (
                <option key={ingredient.id} value={ingredient.id}>
                  {ingredient.name} ({ingredient.unit})
                </option>
              ))}
            </select>

            <input
              type="number"
              placeholder="Quantity"
              value={quantity}
              onChange={(e) => setQuantity(e.target.value)}
              step="0.1"
              min="0.1"
              className="quantity-input"
            />

            <button
              type="button"
              className="add-ingredient-btn"
              onClick={handleAddIngredient}
              disabled={!selectedIngredient}
            >
              Add Ingredient
            </button>
          </div>
        </div>

        <div className="form-actions">
          <button
            type="button"
            onClick={() => navigate("/merchant/products")}
            className="cancel-button"
          >
            Cancel
          </button>
          {isEditing && (
            <button
              type="button"
              className="delete-button"
              onClick={() => handleDelete(id)}
            >
              Delete
            </button>
          )}
          <button type="submit" disabled={submitting} className="submit-button">
            {submitting
              ? "Saving..."
              : isEditing
              ? "Update Product"
              : "Add Product"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default ProductForm; 
