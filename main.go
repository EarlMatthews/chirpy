package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/EarlMatthews/chirpy/internal/auth"
	"github.com/EarlMatthews/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct{
	fileserverHits atomic.Int32
	DB *database.Queries
	platform string
	secret	string
	
}

type Users struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
	Password  string  `json:"password,omitempty"`
	//Expires	string	`json:"expires_in_seconds,omitempty"`
	Token	string	`json:"token,omitempty"`
	RefreshToken string	`json:"refresh_token"`
}

// type UsersNoPassword struct {
// 	ID        string `json:"id"`
// 	CreatedAt string `json:"created_at"`
// 	UpdatedAt string `json:"updated_at"`
// 	Email     string `json:"email"`
// }

type Chirp struct {
	Body string `json:"body"`
	//UserID string `json:"user_id"`
}

type ChirpShown struct {
	ID 			string `json:"id"`
	CreatedAt	string `json:"created_at"`
	UpdatedAt	string `json:"updated_at"`
	Body		string `json:"body"`
	UserID		string `json:"user_id"`
}

type Cleanedchirp struct {
	CleanedBody string `json:"cleaned_body"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}){
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
    respondWithJSON(w, code, map[string]string{"error": msg})
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request){
	// Check if the request method is POST. If not, return without processing.
	if r.Method != http.MethodPost {
		//http.Error(w, "Method is not POST", http.StatusMethodNotAllowed)
		
		return
	}
	// Declare a variable 'user' of type Users to hold the user data from the request body.
	var user Users

	// Decode the JSON body of the request into the 'user' variable.
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w,http.StatusBadRequest, "Invalid JSON body")
		return
	}

	// Declare a variable 'retrUser' of type database.User to store the retrieved user data from the database.
	var retrUser database.User

	// Attempt to log in the user by fetching the user with the given email from the database.
	retrUser, err = cfg.DB.Login(r.Context(),user.Email)
	// If there's an error (e.g., user not found), respond with an Unauthorized status and message.
	if err != nil{
		respondWithError(w,http.StatusUnauthorized, "Incorrect email or password")
		return
	}
	err = auth.CheckPasswordHash(user.Password, retrUser.HashedPassword)
	if err == nil {
    	// // If passwords match, prepare a response object with user details.
    	retrUser.HashedPassword = ""
		userResponse := Users{
			ID:        retrUser.ID.String(),
    		CreatedAt: retrUser.CreatedAt.Time.Format(time.RFC3339), // Formats to ISO-8601
    		UpdatedAt: retrUser.UpdatedAt.Time.Format(time.RFC3339),
    		Email:		retrUser.Email,
			}

		// Generate a JWT token for the authenticated user.
		newToken, err := auth.MakeJWT(retrUser.ID, cfg.secret, time.Duration(1 * int(time.Hour)))
		if err != nil{
			// If there's an error in generating the token, respond with an Unauthorized status and message.
			respondWithError(w,http.StatusUnauthorized, "Bad Token")
			return
		}
		userResponse.Token = newToken
		newRefreshToken, err := auth.MakeRefreshToken()
		if err != nil{
			respondWithError(w,http.StatusUnauthorized,"Bad Refresh Token")
			return
		}
		// Create a new refresh token
		userResponse.RefreshToken = newRefreshToken
		refresh_user_uuid, err := uuid.Parse(userResponse.ID)
		if err != nil{
			respondWithError(w,http.StatusUnauthorized,"Error Creating UUID")
			return
		}
		// store the refresh token in the database
		refreshTokenParams := database.StoreRefreshTokenParams{
			Token: newRefreshToken,
			UserID: uuid.NullUUID{UUID: refresh_user_uuid, Valid: true},
		}
		err = cfg.DB.StoreRefreshToken(r.Context(), refreshTokenParams)
		if err != nil{
			respondWithError(w,http.StatusUnauthorized,"Problem storing Refresh Token")
			return
		}
		// return the user Response
    	respondWithJSON(w, http.StatusOK, userResponse)
	} else {
    	// If passwords don't match, respond with an Unauthorized status and message.
    	respondWithError(w, http.StatusUnauthorized, "Incorrect email or password")
    	return
	}
}

func (cfg *apiConfig) showChirps(w http.ResponseWriter, r *http.Request){
	
	dbChirp, err := cfg.DB.ShowChirpsAll(r.Context())
	if err != nil{
	respondWithError(w,http.StatusBadRequest, "Error Connecting to Database" + err.Error())
	return
	}

	var chirpResponse []ChirpShown
	for _, chirp := range dbChirp {
		chirpResponse = append(chirpResponse, ChirpShown{
			ID:		chirp.ID.String(),
			CreatedAt: chirp.CreatedAt.Time.String(),
			UpdatedAt: chirp.UpdatedAt.Time.String(),
			Body: chirp.Body.String,
			UserID: chirp.UserID.UUID.String(),
		})
	}
	respondWithJSON(w, http.StatusOK,chirpResponse)
}

func (cfg *apiConfig) showOneChirp(w http.ResponseWriter, r *http.Request){
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil{
		respondWithError(w,http.StatusNotFound,"Invalid UUID")
		return
	}

	dbChirp, err := cfg.DB.ShowOneChirp(r.Context(), chirpID)
	if err != nil{
		respondWithError(w,http.StatusNotFound,"Invalid UUID")
		return
	}
	chirpResponse := ChirpShown{
		ID:			dbChirp.ID.String(),
		CreatedAt:	dbChirp.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:	dbChirp.UpdatedAt.Time.UTC().Format(time.RFC3339),
		Body:		dbChirp.Body.String,
		UserID:		dbChirp.UserID.UUID.String(),
	}

	respondWithJSON(w,http.StatusOK,chirpResponse)

}

func (cfg *apiConfig)chirps(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "Method is not POST", http.StatusMethodNotAllowed)
 		return
 	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w,http.StatusUnauthorized,err.Error())
	}
	userid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w,http.StatusUnauthorized,"Bad Token " + err.Error())
	}

	var chirp Chirp
	err = json.NewDecoder(r.Body).Decode(&chirp)
	if err != nil {
 		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
 		return
 	}
	if len(chirp.Body) > 140{
		respondWithError(w,http.StatusBadRequest,"{\"error\":\"Chirp is too long\"}")
	} 
	chirpUserID := userid
	params := database.CreateChirpParams{
        Body:   sql.NullString{String: chirp.Body, Valid: true},
		UserID: uuid.NullUUID{UUID: chirpUserID, Valid: true},
    }

	dbChirp, err := cfg.DB.CreateChirp(r.Context(),params)
	if err != nil{
	respondWithError(w,http.StatusBadRequest, "Error Connecting to Database" + err.Error())
	return
	}

	// chirpResponse := Chirp{
	// 	Body: dbChirp.Body.String,
	// 	UserID: dbChirp.UserID.UUID.String(),
	// }

	chirpResponse := ChirpShown{
		ID:			dbChirp.ID.String(),
		CreatedAt:	dbChirp.CreatedAt.Time.UTC().Format(time.RFC3339),
		UpdatedAt:	dbChirp.UpdatedAt.Time.UTC().Format(time.RFC3339),
		Body:		dbChirp.Body.String,
		UserID:		dbChirp.UserID.UUID.String(),
	}
	respondWithJSON(w, http.StatusCreated,chirpResponse)
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	// Next, write a new middleware method on a *apiConfig that increments the fileserverHits counter every time it's called
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment the fileserverHits counter
		cfg.fileserverHits.Add(1)
		
		// Call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig)createUser (w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		//http.Error(w, "Method is not POST", http.StatusMethodNotAllowed)
		
		return
	}

	var user Users
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		respondWithError(w,http.StatusBadRequest, "Invalid JSON body")
		return
	}
	hashedpassword, err := auth.HashPassword(user.Password)
	if err != nil{
		respondWithError(w,http.StatusBadRequest,"bad hash")
	}
	newUser := database.CreateUserParams{
		Email: user.Email,
		HashedPassword: hashedpassword,
	}
	dbuser, err := cfg.DB.CreateUser(r.Context(), newUser)
	if err != nil{
		respondWithError(w,http.StatusBadRequest, "Error Connecting to Database" + err.Error())
		return
	}

	userResponse := Users{
	ID:        dbuser.ID.String(),
    CreatedAt: dbuser.CreatedAt.Time.Format(time.RFC3339), // Formats to ISO-8601
    UpdatedAt: dbuser.UpdatedAt.Time.Format(time.RFC3339),
    Email:     dbuser.Email,
	}
	respondWithJSON(w,http.StatusCreated,userResponse)
}

func healthz(w http.ResponseWriter, r *http.Request){
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	w.Write([]byte("OK"))
}

func metrics (cfg *apiConfig, w http.ResponseWriter, r *http.Request){
	// Create a new handler that writes the number of requests 
	// that have been counted as plain text in this format to the HTTP response
	count := cfg.fileserverHits.Load()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type","text/plain; charset=utf-8")
	html := fmt.Sprintf(`
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>`, count)
	_, _ = w.Write([]byte(html))
}

func (cfg *apiConfig)reset (w http.ResponseWriter, r *http.Request){
	//Finally, create and register a handler on the /reset path that,
	//when hit, will reset your fileserverHits back to 0
	cfg.fileserverHits.Store(0)

	if cfg.platform == "dev" {
		err := cfg.DB.DeleteUser(r.Context())
		if err != nil{
			respondWithError(w,http.StatusBadRequest, "Error Connecting to Database" + err.Error())
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type","text/plain; charset=utf-8")
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
	
}

func (cfg *apiConfig) refresh(w http.ResponseWriter, r *http.Request){
	// Check if the request method is POST. If not, return without processing.
	if r.Method != http.MethodPost {
		//http.Error(w, "Method is not POST", http.StatusMethodNotAllowed)
		
		return
	}
	// read the refresh token from the header, if no refresh token quit
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil{
		respondWithError(w,http.StatusUnauthorized, "No refresh token found")
		return
	}
	// look up refresh token in the database and get the UUID associated 
	refreshTokenInfo, err := cfg.DB.RetrieveRefreshToken(r.Context(),refreshToken)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized,"Token Not Found")
		return
	}
	if time.Now().After(refreshTokenInfo.ExpiresAt.Time){
		respondWithError(w,http.StatusUnauthorized,"Token Expired")
		return
	}
	if refreshTokenInfo.RevokedAt.Valid {
		respondWithError(w,http.StatusUnauthorized,"Token Revoked")
		return
	}
	 accessToken, err := auth.MakeJWT(refreshTokenInfo.UserID.UUID, cfg.secret, time.Hour)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Couldn't create access token")
        return
    }
	//cfg.DB.StoreRefreshToken(r.Context(),refreshTokenParams)
	respondWithJSON(w,http.StatusOK,map[string]string{"token": accessToken})

}

func (cfg *apiConfig) revoke(w http.ResponseWriter, r *http.Request){
	// Check if the request method is POST. If not, return without processing.
	if r.Method != http.MethodPost {
		//http.Error(w, "Method is not POST", http.StatusMethodNotAllowed)
		
		return
	}
	// read the refresh token from the header, if no refresh token quit
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil{
		respondWithError(w,http.StatusUnauthorized, "No refresh token found")
		return
	}
	// look up refresh token in the database and get the UUID associated 
	// refreshTokenInfo, err := cfg.DB.RetrieveRefreshToken(r.Context(),refreshToken)
	// if err != nil{
	// 	respondWithError(w, http.StatusUnauthorized,"Token Not Found")
	// 	return
	// }
	err = cfg.DB.RevokeRefreshToken(r.Context(),refreshToken)
	if err != nil{
		respondWithError(w,http.StatusUnauthorized,"Error revoking token")
	}
	respondWithJSON(w,http.StatusNoContent,nil)
}

func (cfg *apiConfig) updateuser(w http.ResponseWriter, r *http.Request){
	// Check if the request method is PUT. If not, return without processing.
	if r.Method != http.MethodPut {
		//http.Error(w, "Method is not PUT", http.StatusMethodNotAllowed)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w,http.StatusUnauthorized,err.Error())
		return
	}
	userid, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w,http.StatusUnauthorized,"Bad Token " + err.Error())
	}
	// check body to see if both email and password are populated
	type UserInfo struct{
		email string 
		password string
	}
	var ui UserInfo
	err = json.NewDecoder(r.Body).Decode(&ui)
		if err != nil{
			respondWithError(w,http.StatusUnauthorized, err.Error())
		}
	ui.password, err = auth.HashPassword(ui.password)
	if err != nil{
		respondWithError(w, http.StatusUnauthorized, err.Error())
	}
	// update email and password of user in database
	params := database.UpdateAuthParams{
		ID:		userid,
		Email: 	ui.email,
		HashedPassword: ui.password,
	}
	
	user, err := cfg.DB.UpdateAuth(r.Context(),params)
	if err != nil{
		respondWithError(w,http.StatusUnauthorized,err.Error())
		return
	}
	respondWithJSON(w,http.StatusOK,user)

}

func main(){
	err := godotenv.Load()
	if err != nil {
    	log.Fatalf("Error loading .env file")
	}
	dbURL := os.Getenv("DB_URL")


	db, err := sql.Open("postgres", dbURL)
	if err != nil{
		panic(fmt.Sprintf("Failed to connect to the database: %v", err))
	}
	defer db.Close()

	dbQueries := database.New(db)
	mux := http.NewServeMux()
	cfg := &apiConfig{fileserverHits: atomic.Int32{}, DB: dbQueries, platform: os.Getenv("PLATFORM"), secret: os.Getenv("SECRET")}
	// Create a New server
	srv := http.Server{
		Addr: ":8888",
		Handler: mux,
	}

	// Use the http.NewServeMux's .Handle() method to add a handler for the root path (/).
	
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(".")))))
	mux.Handle("/app/assets/", cfg.middlewareMetricsInc(http.StripPrefix("/app/assets/", http.FileServer(http.Dir("assets")))))
	mux.HandleFunc("GET /admin/healthz",healthz)
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics(cfg, w, r)
	} )
	mux.HandleFunc("POST /admin/reset", cfg.reset)
	//mux.HandleFunc("POST /api/validate_chirp", validateChirp)
	mux.HandleFunc("POST /api/chirps", cfg.chirps)
	mux.HandleFunc("POST /api/users", cfg.createUser)
	mux.HandleFunc("POST /api/login",cfg.login)
	mux.HandleFunc("POST /api/refresh", cfg.refresh)
	mux.HandleFunc("POST /api/revoke", cfg.revoke)
	mux.HandleFunc("PUT /api/users", cfg.updateuser)
	mux.HandleFunc("GET /api/chirps", cfg.showChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.showOneChirp)
	//Starting the new server
	if err := srv.ListenAndServe(); err != nil {
        panic(err)
    }
}
