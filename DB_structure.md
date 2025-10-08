# Database structure
**Table `feeds`**  
Column `id` (primary int): Unique ID  
Column `title` (string): RSS title  
Column `description` (string): RSS description  
Column `link` (nullable string): Link to HTML feed
Column `fetchFrom` (nullable string): Link to RSS feed
Column `language` (nullable string): RSS language  
Column `ttl` (nullable int): RSS time-to-live (cache time before refreshing) in minutes  

**Table `items`**  
Column `id` (primary int): Unique ID  
Column `feed_id` (foreign int): References `feeds.id`  
Column `guid` (unique nullable string): Unique RSS item identifier  
Column `title` (nullable string): Item title  
Column `description` (nullable string): Item description  
Column `link` (nullable string): Item URL  
Column `author` (nullable string): Author of the item  
Column `pubDate` (nullable datetime): Publication date/time  
Column `read` (boolean, default false): Whether the item has been marked as read  
Column `enclosure_url` (nullable string): URL for media enclosure (if any)  
Column `enclosure_type` (nullable string): MIME type of the enclosure  
Column `enclosure_length` (nullable int): Length in bytes of the enclosure  
