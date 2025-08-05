# Use a slim Node.js base image
FROM node:22-slim

# Install necessary dependencies for Chromium to run
RUN apt-get update && \
    apt-get install -y wget ca-certificates fonts-liberation libappindicator3-1 libasound2 libatk-bridge2.0-0 \
    libatk1.0-0 libcups2 libdbus-1-3 libgdk-pixbuf2.0-0 libnspr4 libnss3 libx11-xcb1 libxcomposite1 \
    libxdamage1 libxrandr2 xdg-utils --no-install-recommends && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy package files and install dependencies
COPY package*.json ./
RUN npm install

# Install Puppeteer Chrome browser
RUN npx puppeteer browsers install chrome

# Copy the rest of the app source
COPY . .

# Set environment variables
ENV PORT=3001

# Expose port
EXPOSE 3001

# Start the application
CMD ["node", "index.js"]