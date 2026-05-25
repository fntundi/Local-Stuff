# Sentinel NOC - Frontend Dockerfile
# ===================================
# This Dockerfile builds the React frontend service.
#
# Build:
#   docker build -t sentinel-frontend ./frontend
#
# Run:
#   docker run -p 3000:3000 sentinel-frontend

# Use Node.js 18 Alpine for smaller image
FROM node:20-alpine

# Set environment variables
ENV NODE_ENV=development \
    CHOKIDAR_USEPOLLING=true

# Set working directory
WORKDIR /app

# Install dependencies for node-gyp (needed for some packages)
RUN apk add --no-cache python3 make g++

# yarn.lock is optional: copied when present, generated when absent.
# Once yarn.lock is committed, restore --frozen-lockfile for reproducible CI builds.
COPY package.json yarn.loc[k] ./

# Install dependencies
RUN yarn install --network-timeout 600000

# Copy application code
COPY . .

# Build argument for backend URL
ARG REACT_APP_BACKEND_URL=http://localhost:8001
ENV REACT_APP_BACKEND_URL=$REACT_APP_BACKEND_URL

# Expose the port
EXPOSE 3000

# Start the development server
CMD ["yarn", "start"]
