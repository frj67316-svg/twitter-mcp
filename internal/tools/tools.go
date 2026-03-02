// Copyright 2024 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tools

import (
	"twitter-mcp/internal/globals"
	"twitter-mcp/internal/middlewares"
	"twitter-mcp/internal/schedule"
	"twitter-mcp/internal/twitter"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolsManagerDependencies struct {
	AppCtx *globals.ApplicationContext

	McpServer     *server.MCPServer
	Middlewares   []middlewares.ToolMiddleware
	TwitterClient *twitter.Client
	ScheduleStore *schedule.Store
}

type ToolsManager struct {
	dependencies ToolsManagerDependencies
	toolPrefix   string
}

func NewToolsManager(deps ToolsManagerDependencies) *ToolsManager {
	return &ToolsManager{
		dependencies: deps,
		toolPrefix:   deps.AppCtx.ToolPrefix,
	}
}

func (tm *ToolsManager) toolName(base string) string {
	return tm.toolPrefix + base
}

// wrapWithMiddlewares applies all configured middlewares to a tool handler
func (tm *ToolsManager) wrapWithMiddlewares(handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	// Apply middlewares in reverse order so the first middleware in the list is the outermost
	for i := len(tm.dependencies.Middlewares) - 1; i >= 0; i-- {
		handler = tm.dependencies.Middlewares[i].Middleware(handler)
	}
	return handler
}

func (tm *ToolsManager) AddTools() {
	// post_tweet - Post a new tweet
	tool := mcp.NewTool(tm.toolName("post_tweet"),
		mcp.WithDescription("Post a new tweet to Twitter/X"),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("The text content of the tweet (max 280 characters)"),
		),
		mcp.WithString("reply_to_id",
			mcp.Description("Optional: Tweet ID to reply to"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolPostTweet))

	// delete_tweet - Delete a tweet
	tool = mcp.NewTool(tm.toolName("delete_tweet"),
		mcp.WithDescription("Delete a tweet by its ID"),
		mcp.WithString("tweet_id",
			mcp.Required(),
			mcp.Description("The ID of the tweet to delete"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolDeleteTweet))

	// get_timeline - Get home timeline
	tool = mcp.NewTool(tm.toolName("get_timeline"),
		mcp.WithDescription("Get the authenticated user's home timeline (recent tweets from followed accounts)"),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of tweets to return (default: 10, max: 100)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetTimeline))

	// get_mentions - Get mentions
	tool = mcp.NewTool(tm.toolName("get_mentions"),
		mcp.WithDescription("Get tweets that mention the authenticated user"),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of mentions to return (default: 10, max: 100)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetMentions))

	// search_tweets - Search for tweets
	tool = mcp.NewTool(tm.toolName("search_tweets"),
		mcp.WithDescription("Search for tweets matching a query. Supports Twitter search operators."),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query (e.g., 'kubernetes', 'from:user', '#hashtag')"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of tweets to return (default: 10, max: 100)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolSearchTweets))

	// get_trends - Get trending topics
	tool = mcp.NewTool(tm.toolName("get_trends"),
		mcp.WithDescription("Get trending topics for a location. Use WOEID: 1=Worldwide, 23424950=Spain, 23424977=USA, 766273=Madrid"),
		mcp.WithNumber("woeid",
			mcp.Description("Where On Earth ID for location (default: 1 = Worldwide)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetTrends))

	// search_topics - Search for content across multiple topics
	tool = mcp.NewTool(tm.toolName("search_topics"),
		mcp.WithDescription("Search for trending content across multiple topics at once. Useful for exploring what's being discussed about specific subjects."),
		mcp.WithArray("topics",
			mcp.Required(),
			mcp.Description("Array of topics/queries to search for (e.g., ['kubernetes', 'AI news', 'golang OR rust'])"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of tweets per topic (default: 5, max: 20)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolSearchTopics))

	// get_topics_heat - Get heat/popularity score for topics
	tool = mcp.NewTool(tm.toolName("get_topics_heat"),
		mcp.WithDescription("Analyze how 'hot' multiple topics are on Twitter. Returns a heat score (0-100) based on tweet volume and engagement metrics (likes, retweets, replies). Useful for comparing topic popularity."),
		mcp.WithArray("topics",
			mcp.Required(),
			mcp.Description("Array of topics to analyze (e.g., ['kubernetes', 'docker', 'podman'])"),
		),
		mcp.WithNumber("sample_size",
			mcp.Description("Number of tweets to sample per topic for analysis (default: 20, max: 100)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetTopicsHeat))

	// get_me - Get authenticated user info
	tool = mcp.NewTool(tm.toolName("get_me"),
		mcp.WithDescription("Get information about the authenticated Twitter user"),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetMe))

	// like_tweet - Like a tweet
	tool = mcp.NewTool(tm.toolName("like_tweet"),
		mcp.WithDescription("Like a tweet"),
		mcp.WithString("tweet_id",
			mcp.Required(),
			mcp.Description("The ID of the tweet to like"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolLikeTweet))

	// unlike_tweet - Remove like from a tweet
	tool = mcp.NewTool(tm.toolName("unlike_tweet"),
		mcp.WithDescription("Remove like from a tweet"),
		mcp.WithString("tweet_id",
			mcp.Required(),
			mcp.Description("The ID of the tweet to unlike"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolUnlikeTweet))

	// retweet - Retweet a tweet
	tool = mcp.NewTool(tm.toolName("retweet"),
		mcp.WithDescription("Retweet a tweet"),
		mcp.WithString("tweet_id",
			mcp.Required(),
			mcp.Description("The ID of the tweet to retweet"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolRetweet))

	// undo_retweet - Remove a retweet
	tool = mcp.NewTool(tm.toolName("undo_retweet"),
		mcp.WithDescription("Remove a retweet"),
		mcp.WithString("tweet_id",
			mcp.Required(),
			mcp.Description("The ID of the tweet to un-retweet"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolUndoRetweet))

	// follow_user - Follow a user
	tool = mcp.NewTool(tm.toolName("follow_user"),
		mcp.WithDescription("Follow a Twitter user"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the user to follow (without @)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolFollowUser))

	// unfollow_user - Unfollow a user
	tool = mcp.NewTool(tm.toolName("unfollow_user"),
		mcp.WithDescription("Unfollow a Twitter user"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the user to unfollow (without @)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolUnfollowUser))

	// get_user_profile - Get a user's profile
	tool = mcp.NewTool(tm.toolName("get_user_profile"),
		mcp.WithDescription("Get a Twitter user's profile information including bio, followers count, etc."),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the user (without @)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetUserProfile))

	// get_user_tweets - Get a user's recent tweets
	tool = mcp.NewTool(tm.toolName("get_user_tweets"),
		mcp.WithDescription("Get recent tweets from a specific user"),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("The username of the user (without @)"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of tweets to return (default: 10, max: 100)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetUserTweets))

	// bookmark_tweet - Bookmark a tweet
	tool = mcp.NewTool(tm.toolName("bookmark_tweet"),
		mcp.WithDescription("Bookmark a tweet for later"),
		mcp.WithString("tweet_id",
			mcp.Required(),
			mcp.Description("The ID of the tweet to bookmark"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolBookmarkTweet))

	// remove_bookmark - Remove a bookmark
	tool = mcp.NewTool(tm.toolName("remove_bookmark"),
		mcp.WithDescription("Remove a bookmark from a tweet"),
		mcp.WithString("tweet_id",
			mcp.Required(),
			mcp.Description("The ID of the tweet to remove from bookmarks"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolRemoveBookmark))

	// get_bookmarks - Get bookmarked tweets
	tool = mcp.NewTool(tm.toolName("get_bookmarks"),
		mcp.WithDescription("Get your bookmarked tweets"),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of bookmarks to return (default: 10, max: 100)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolGetBookmarks))

	// post_thread - Post a thread of tweets
	tool = mcp.NewTool(tm.toolName("post_thread"),
		mcp.WithDescription("Post a thread (multiple connected tweets)"),
		mcp.WithArray("tweets",
			mcp.Required(),
			mcp.Description("Array of tweet texts to post as a thread (first tweet is the head)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolPostThread))

	// schedule_tweet - Schedule a tweet or thread
	tool = mcp.NewTool(tm.toolName("schedule_tweet"),
		mcp.WithDescription("Schedule a tweet or thread for later publishing. Content is always an array of strings (one element for a tweet, multiple for a thread)."),
		mcp.WithString("type",
			mcp.Required(),
			mcp.Description("Type of content: 'tweet' or 'thread'"),
		),
		mcp.WithArray("content",
			mcp.Required(),
			mcp.Description("Array of strings. One element for a tweet, multiple for a thread."),
		),
		mcp.WithString("scheduled_at",
			mcp.Required(),
			mcp.Description("Date and time to publish, in RFC3339 format (e.g. 2026-02-25T10:00:00Z)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolScheduleTweet))

	// schedule_update - Update a scheduled tweet
	tool = mcp.NewTool(tm.toolName("schedule_update"),
		mcp.WithDescription("Update a scheduled tweet or thread. Only provided fields will be updated."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the scheduled tweet to update"),
		),
		mcp.WithString("type",
			mcp.Description("Type of content: 'tweet' or 'thread'"),
		),
		mcp.WithArray("content",
			mcp.Description("New content array"),
		),
		mcp.WithString("scheduled_at",
			mcp.Description("New scheduled date in RFC3339 format"),
		),
		mcp.WithBoolean("reviewed",
			mcp.Description("Mark as reviewed (true) or back to pending (false)"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolScheduleUpdate))

	// schedule_delete - Delete a scheduled tweet
	tool = mcp.NewTool(tm.toolName("schedule_delete"),
		mcp.WithDescription("Delete a scheduled tweet by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the scheduled tweet to delete"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolScheduleDelete))

	// schedule_list - List scheduled tweets
	tool = mcp.NewTool(tm.toolName("schedule_list"),
		mcp.WithDescription("List scheduled tweets, optionally filtered by status"),
		mcp.WithString("status",
			mcp.Description("Filter by status: 'pending', 'reviewed', 'published', 'failed'. Leave empty for all."),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolScheduleList))

	// schedule_get_publishable - Get tweets ready to publish
	tool = mcp.NewTool(tm.toolName("schedule_get_publishable"),
		mcp.WithDescription("Get scheduled tweets that are ready to publish: reviewed, scheduled time is past, and enough time has passed since the last published tweet."),
		mcp.WithNumber("min_hours_since_last",
			mcp.Description("Minimum hours since last published tweet (default: 1). Use 0 to ignore."),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolScheduleGetPublishable))

	// schedule_publish - Publish a scheduled tweet
	tool = mcp.NewTool(tm.toolName("schedule_publish"),
		mcp.WithDescription("Publish a specific scheduled tweet or thread by ID"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("ID of the scheduled tweet to publish"),
		),
	)
	tm.dependencies.McpServer.AddTool(tool, tm.wrapWithMiddlewares(tm.HandleToolSchedulePublish))
}
