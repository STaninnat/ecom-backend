package router

import (
	"net/http"

	"time"

	"github.com/STaninnat/ecom-backend/handlers"
	authhandlers "github.com/STaninnat/ecom-backend/handlers/auth"
	carthandlers "github.com/STaninnat/ecom-backend/handlers/cart"
	categoryhandlers "github.com/STaninnat/ecom-backend/handlers/category"
	orderhandlers "github.com/STaninnat/ecom-backend/handlers/order"
	paymenthandlers "github.com/STaninnat/ecom-backend/handlers/payment"
	producthandlers "github.com/STaninnat/ecom-backend/handlers/product"
	reviewhandlers "github.com/STaninnat/ecom-backend/handlers/review"
	uploadhandlers "github.com/STaninnat/ecom-backend/handlers/upload"
	userhandlers "github.com/STaninnat/ecom-backend/handlers/user"
	intmongo "github.com/STaninnat/ecom-backend/internal/mongo"
	"github.com/STaninnat/ecom-backend/middlewares"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"
)

// RouterConfig holds the configuration for setting up the API router.
type RouterConfig struct {
	*handlers.HandlersConfig
}

// SetupRouter initializes and returns the main chi.Mux router for the API.
// Sets up global middleware, static file serving, and mounts the versioned API subrouter.
// Organizes resource-specific subrouters for clarity and maintainability.
func (apicfg *RouterConfig) SetupRouter(logger *logrus.Logger) *chi.Mux {
	// --- Global Middleware Setup ---
	// Create the main router instance
	router := chi.NewRouter()

	// Standard logging for all requests
	router.Use(middleware.Logger)
	// Recover from panics and return 500 errors
	router.Use(middleware.Recoverer)

	// Security headers for all responses (custom middleware)
	router.Use(middlewares.SecurityHeaders)
	// Attach a unique request ID to each request (custom middleware)
	router.Use(middlewares.RequestIDMiddleware)
	// Custom logging middleware with path-based filtering:
	// - Only logs requests to /v1 and its subpaths
	// - Skips logging for /v1/healthz and /v1/error endpoints
	router.Use(middlewares.LoggingMiddleware(
		logger,
		map[string]struct{}{"/v1": {}},
		map[string]struct{}{"/v1/healthz": {}, "/v1/error": {}},
	))

	// Add distributed rate limiting middleware (100 requests per 15 minutes per IP, backed by Redis)
	router.Use(middlewares.RedisRateLimiter(apicfg.RedisClient, 100, 15*time.Minute))

	// CORS middleware: allows cross-origin requests from any HTTP/HTTPS origin.
	// - AllowedOrigins: Accepts all subdomains for both http and https (useful for dev and prod)
	// - AllowedMethods: Permits common RESTful methods
	// - AllowedHeaders: Accepts any header (for flexibility with clients)
	// - ExposedHeaders: Exposes 'Link' header to clients
	// - AllowCredentials: Disallows cookies/auth headers by default (set to true if needed)
	// - MaxAge: Caches preflight response for 5 minutes
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	useS3 := apicfg.HandlersConfig.UploadBackend == "s3"
	uploadPath := apicfg.HandlersConfig.UploadPath

	// --- Static File Server ---
	// Serve static files from the same directory as uploads (uploadPath).
	// This ensures that files uploaded via the local backend are accessible at /static/*.
	// If using S3, this will only serve files present in the local uploadPath directory.
	//
	// Why this matters:
	// - If uploadPath and the static file server path differ, uploaded files won't be accessible at /static/*.
	// - If using S3, files uploaded to S3 will NOT be accessible at /static/* unless you sync them to the local uploadPath or implement an S3 proxy handler.
	//
	// Example:
	//   uploadPath = "./uploads"  -->  /static/* serves uploaded files (local backend)
	//   uploadPath = "/var/data/uploads"  -->  /static/* serves from that directory
	//   S3 backend  -->  /static/* only serves files present in uploadPath, not S3
	fs := http.FileServer(http.Dir(uploadPath))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	// --- Handler Configurations ---
	// Auth handler config: provides dependencies for auth-related handlers
	authHandlersConfig := &authhandlers.HandlersAuthConfig{HandlersConfig: apicfg.HandlersConfig}
	// User handler config: provides dependencies for user-related handlers
	userHandlersConfig := &userhandlers.HandlersUserConfig{HandlersConfig: apicfg.HandlersConfig}
	// Product handler config: includes DB, connection, and logger for product endpoints
	productHandlersConfig := &producthandlers.HandlersProductConfig{
		DB:     apicfg.DB,
		DBConn: apicfg.DBConn,
		Logger: apicfg.HandlersConfig,
	}
	// Category handler config: for category endpoints
	categoryHandlersConfig := &categoryhandlers.HandlersCategoryConfig{HandlersConfig: apicfg.HandlersConfig}

	// --- Upload Handler Setup ---
	// Set up the upload handler and service, supporting both S3 and local backends
	productDB := uploadhandlers.NewProductDBAdapter(apicfg.HandlersConfig.DB)
	var fileStorage uploadhandlers.FileStorage
	if useS3 {
		// Use S3 for file storage
		fileStorage = &uploadhandlers.S3FileStorage{
			S3Client:   apicfg.HandlersConfig.S3Client, // AWS S3 client
			BucketName: apicfg.HandlersConfig.S3Bucket, // S3 bucket name
		}
	} else {
		// Use local filesystem for file storage
		fileStorage = &uploadhandlers.LocalFileStorage{}
	}
	// Upload service combines DB, path, and storage backend
	uploadService := uploadhandlers.NewUploadService(productDB, uploadPath, fileStorage)
	// Upload handler config: provides dependencies for upload endpoints
	uploadHandlersConfig := &uploadhandlers.HandlersUploadConfig{
		HandlersConfig: apicfg.HandlersConfig,
		Logger:         apicfg.HandlersConfig,
		UploadPath:     uploadPath,
		Service:        uploadService,
	}

	// --- Order and Payment Handler Configs ---
	orderHandlersConfig := &orderhandlers.HandlersOrderConfig{HandlersConfig: apicfg.HandlersConfig}
	paymentHandlersConfig := &paymenthandlers.HandlersPaymentConfig{HandlersConfig: apicfg.HandlersConfig}

	// --- Review and Cart Service Setup ---
	var reviewHandlersConfig *reviewhandlers.HandlersReviewConfig
	var cartHandlersConfig *carthandlers.HandlersCartConfig
	if apicfg.MongoDB != nil {
		reviewMongoRepo := intmongo.NewReviewMongo(apicfg.MongoDB)
		cartMongoRepo := intmongo.NewCartMongo(apicfg.MongoDB)

		// Review handler config and service
		reviewHandlersConfig = &reviewhandlers.HandlersReviewConfig{
			HandlersConfig: apicfg.HandlersConfig,
		}
		reviewService := reviewhandlers.NewReviewService(reviewMongoRepo)
		err := reviewHandlersConfig.InitReviewService(reviewService)
		if err != nil {
			logger.Fatal("Failed to initialize review service:", err)
		}

		// Cart handler config and service
		cartHandlersConfig = &carthandlers.HandlersCartConfig{
			HandlersConfig: apicfg.HandlersConfig,
		}
		cartService := carthandlers.NewCartServiceWithDeps(
			cartMongoRepo,
			apicfg.DB,
			apicfg.DBConn,
			apicfg.RedisClient,
		)
		if err := cartHandlersConfig.InitCartService(cartService); err != nil {
			logger.Fatal("Failed to initialize cart service:", err)
		}
	}

	// --- Cache Configurations ---
	// Add caching for read-heavy endpoints
	productsCacheConfig := middlewares.CacheConfig{
		TTL:          30 * time.Minute, // Cache products for 30 minutes
		KeyPrefix:    "products",
		CacheService: apicfg.CacheService,
	}

	categoriesCacheConfig := middlewares.CacheConfig{
		TTL:          1 * time.Hour, // Cache categories for 1 hour
		KeyPrefix:    "categories",
		CacheService: apicfg.CacheService,
	}

	v1Router := chi.NewRouter()

	// --- Health and Error Endpoints ---
	v1Router.Get("/healthz", Adapt(handlers.HandlerReadiness)) // Health check endpoint
	v1Router.Get("/errorz", Adapt(handlers.HandlerError))      // Error simulation endpoint

	// --- Auth Subrouter ---
	authRouter := chi.NewRouter()
	authRouter.Post("/signup", middlewares.NoCacheHeaders(Adapt(authHandlersConfig.HandlerSignUp)).(http.HandlerFunc))        // User registration
	authRouter.Post("/signin", middlewares.NoCacheHeaders(Adapt(authHandlersConfig.HandlerSignIn)).(http.HandlerFunc))        // User login
	authRouter.Post("/signout", middlewares.NoCacheHeaders(Adapt(authHandlersConfig.HandlerSignOut)).(http.HandlerFunc))      // User logout
	authRouter.Post("/refresh", middlewares.NoCacheHeaders(Adapt(authHandlersConfig.HandlerRefreshToken)).(http.HandlerFunc)) // Refresh JWT tokens
	authRouter.Get("/google/signin", Adapt(authHandlersConfig.HandlerGoogleSignIn))                                           // Google OAuth2 start
	authRouter.Get("/google/callback", Adapt(authHandlersConfig.HandlerGoogleCallback))                                       // Google OAuth2 callback
	v1Router.Mount("/auth", authRouter)

	// --- User Subrouter ---
	usersRouter := chi.NewRouter()
	usersRouter.Get("/", middlewares.NoCacheHeaders(WithUser(userHandlersConfig.AuthHandlerGetUser)).(http.HandlerFunc))    // Get current user profile
	usersRouter.Put("/", middlewares.NoCacheHeaders(WithUser(userHandlersConfig.AuthHandlerUpdateUser)).(http.HandlerFunc)) // Update user profile
	v1Router.Mount("/users", usersRouter)

	// --- Product Subrouter ---
	productsRouter := chi.NewRouter()
	productsRouter.Get("/", middlewares.CacheMiddleware(productsCacheConfig)(WithOptionalUser(productHandlersConfig.HandlerGetAllProducts)).(http.HandlerFunc))              // List all products (cached)
	productsRouter.Get("/filter", middlewares.CacheMiddleware(productsCacheConfig)(WithOptionalUser(productHandlersConfig.HandlerFilterProducts)).(http.HandlerFunc))        // Filter products (cached)
	productsRouter.Get("/{id}", WithUser(productHandlersConfig.HandlerGetProductByID))                                                                                       // Get product details (requires auth)
	productsRouter.Post("/", middlewares.InvalidateCache(apicfg.CacheService, "products:*")(WithAdmin(productHandlersConfig.HandlerCreateProduct)).(http.HandlerFunc))       // Admin: create product, invalidates cache
	productsRouter.Put("/", middlewares.InvalidateCache(apicfg.CacheService, "products:*")(WithAdmin(productHandlersConfig.HandlerUpdateProduct)).(http.HandlerFunc))        // Admin: update product, invalidates cache
	productsRouter.Delete("/{id}", middlewares.InvalidateCache(apicfg.CacheService, "products:*")(WithAdmin(productHandlersConfig.HandlerDeleteProduct)).(http.HandlerFunc)) // Admin: delete product, invalidates cache
	productsRouter.Post("/upload-image", WithAdmin(uploadHandlersConfig.HandlerUploadProductImage))                                                                          // Admin: upload product image
	productsRouter.Post("/{id}/image", WithAdmin(uploadHandlersConfig.HandlerUpdateProductImageByID))                                                                        // Admin: update product image
	v1Router.Mount("/products", productsRouter)

	// --- Category Subrouter ---
	categoriesRouter := chi.NewRouter()
	categoriesRouter.Get("/", middlewares.CacheMiddleware(categoriesCacheConfig)(WithOptionalUser(categoryHandlersConfig.HandlerGetAllCategories)).(http.HandlerFunc))             // List all categories (cached)
	categoriesRouter.Post("/", middlewares.InvalidateCache(apicfg.CacheService, "categories:*")(WithAdmin(categoryHandlersConfig.HandlerCreateCategory)).(http.HandlerFunc))       // Admin: create category, invalidates cache
	categoriesRouter.Put("/", middlewares.InvalidateCache(apicfg.CacheService, "categories:*")(WithAdmin(categoryHandlersConfig.HandlerUpdateCategory)).(http.HandlerFunc))        // Admin: update category, invalidates cache
	categoriesRouter.Delete("/{id}", middlewares.InvalidateCache(apicfg.CacheService, "categories:*")(WithAdmin(categoryHandlersConfig.HandlerDeleteCategory)).(http.HandlerFunc)) // Admin: delete category, invalidates cache
	v1Router.Mount("/categories", categoriesRouter)

	// --- Order Subrouter ---
	ordersRouter := chi.NewRouter()
	ordersRouter.Post("/", WithUser(orderHandlersConfig.HandlerCreateOrder))                           // Create new order
	ordersRouter.Get("/user", WithUser(orderHandlersConfig.HandlerGetUserOrders))                      // Get orders for current user
	ordersRouter.Get("/items/{order_id}", WithUser(orderHandlersConfig.HandlerGetOrderItemsByOrderID)) // Get items for a specific order
	ordersRouter.Put("/{order_id}/status", WithAdmin(orderHandlersConfig.HandlerUpdateOrderStatus))    // Admin: update order status
	ordersRouter.Delete("/{order_id}", WithAdmin(orderHandlersConfig.HandlerDeleteOrder))              // Admin: delete order
	ordersRouter.Get("/", WithAdmin(orderHandlersConfig.HandlerGetAllOrders))                          // Admin: list all orders
	v1Router.Mount("/orders", ordersRouter)

	// --- Cart Subrouter ---
	if cartHandlersConfig != nil {
		cartRouter := chi.NewRouter()
		cartRouter.Post("/items", WithUser(cartHandlersConfig.HandlerAddItemToUserCart))        // Add item to user cart
		cartRouter.Put("/items", WithUser(cartHandlersConfig.HandlerUpdateItemQuantity))        // Update item quantity in user cart
		cartRouter.Get("/items", WithUser(cartHandlersConfig.HandlerGetUserCart))               // Get current user's cart
		cartRouter.Delete("/items", WithUser(cartHandlersConfig.HandlerRemoveItemFromUserCart)) // Remove item from user cart
		cartRouter.Delete("/", WithUser(cartHandlersConfig.HandlerClearUserCart))               // Clear user cart
		cartRouter.Post("/checkout", WithUser(cartHandlersConfig.HandlerCheckoutUserCart))      // Checkout user cart
		v1Router.Mount("/cart", cartRouter)
	}
	// --- Guest Cart Subrouter ---
	if cartHandlersConfig != nil {
		guestCartRouter := chi.NewRouter()
		guestCartRouter.Post("/items", Adapt(cartHandlersConfig.HandlerAddItemToGuestCart))        // Add item to guest cart (no auth)
		guestCartRouter.Get("/", Adapt(cartHandlersConfig.HandlerGetGuestCart))                    // Get guest cart (no auth)
		guestCartRouter.Put("/items", Adapt(cartHandlersConfig.HandlerUpdateGuestItemQuantity))    // Update item in guest cart (no auth)
		guestCartRouter.Delete("/items", Adapt(cartHandlersConfig.HandlerRemoveItemFromGuestCart)) // Remove item from guest cart (no auth)
		guestCartRouter.Delete("/", Adapt(cartHandlersConfig.HandlerClearGuestCart))               // Clear guest cart (no auth)
		v1Router.Mount("/guest-cart", guestCartRouter)
	}

	// --- Payment Subrouter ---
	paymentsRouter := chi.NewRouter()
	paymentsRouter.Post("/webhook", Adapt(paymentHandlersConfig.HandlerStripeWebhook))              // Stripe webhook endpoint
	paymentsRouter.Post("/intent", WithUser(paymentHandlersConfig.HandlerCreatePayment))            // Create payment intent
	paymentsRouter.Post("/confirm", WithUser(paymentHandlersConfig.HandlerConfirmPayment))          // Confirm payment
	paymentsRouter.Get("/{order_id}", WithUser(paymentHandlersConfig.HandlerGetPayment))            // Get payment for order
	paymentsRouter.Get("/history", WithUser(paymentHandlersConfig.HandlerGetPaymentHistory))        // Get payment history for user
	paymentsRouter.Post("/{order_id}/refund", WithUser(paymentHandlersConfig.HandlerRefundPayment)) // Refund payment for order
	paymentsRouter.Get("/admin/{status}", WithAdmin(paymentHandlersConfig.HandlerAdminGetPayments)) // Admin: get payments by status
	v1Router.Mount("/payments", paymentsRouter)

	// --- Review Subrouter ---
	if reviewHandlersConfig != nil {
		reviewsRouter := chi.NewRouter()
		reviewsRouter.Get("/product/{product_id}", Adapt(reviewHandlersConfig.HandlerGetReviewsByProductID)) // Get reviews for a product
		reviewsRouter.Get("/{id}", Adapt(reviewHandlersConfig.HandlerGetReviewByID))                         // Get review by ID
		reviewsRouter.Post("/", WithUser(reviewHandlersConfig.HandlerCreateReview))                          // Create review (auth required)
		reviewsRouter.Get("/user", WithUser(reviewHandlersConfig.HandlerGetReviewsByUserID))                 // Get reviews by user
		reviewsRouter.Put("/{id}", WithUser(reviewHandlersConfig.HandlerUpdateReviewByID))                   // Update review (auth required)
		reviewsRouter.Delete("/{id}", WithUser(reviewHandlersConfig.HandlerDeleteReviewByID))                // Delete review (auth required)
		v1Router.Mount("/reviews", reviewsRouter)
	}

	// --- Admin Subrouter ---
	adminRouter := chi.NewRouter()
	adminRouter.Post("/user/promote", WithAdmin(userHandlersConfig.AuthHandlerPromoteUserToAdmin)) // Promote user to admin
	v1Router.Mount("/admin", adminRouter)

	router.Mount("/v1", v1Router)
	return router
}
