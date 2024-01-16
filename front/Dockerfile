# Use the Node.js base image
FROM node:21.5.0-bookworm

# Set the working directory inside the container
WORKDIR /app

# Copy the package*.json files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy all other project files
COPY . .

# Determine the port that will be used by the container
EXPOSE 3003

# Define the command to start the application server
CMD ["npm", "start"]