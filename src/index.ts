import * as Discord from 'discord.js';
import * as dotenv from 'dotenv';
import * as Router from './router';

dotenv.config();

const client = new Discord.Client();
const APIKEY = process.env.API_KEY;

// Inform the operator that the bot is connected
client.on('ready', () => {
  console.log('Bot connected');
});

// Add a ping route to the router
Router.addRoute("ping", "Pong!", (message: Discord.Message, args: string[]) => {
  message.channel.send("Pong!")
});

// Use the helper function to print the help message
Router.addRoute("help", "Help command.", Router.help);

// Handle a message being recieved
client.on('message', message => {
  Router.route(message);
});

client.login(APIKEY); // haha get pranked API key scraper!
