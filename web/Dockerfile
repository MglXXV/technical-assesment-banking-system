# Use a lightweight Node.js image
FROM node:20-alpine

WORKDIR /app

# Copy dependency files first (to leverage Docker's cache)
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy all the frontend source code
COPY . .

# Expose Vite's default port
EXPOSE 5173

# Command to start the development server. 
# The "--host" flag is REQUIRED in Docker for Vite to be accessible from your browser.
CMD ["npm", "run", "dev", "--", "--host"]