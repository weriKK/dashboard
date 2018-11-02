'use struct';

import { myConfig } from "./config.js";

function Column(props) {
	return (
    <li className={"column column-" + props.color}>
			<FeedBoxList id={props.id} feeds={props.feeds} />
		</li>
  );
}

function FeedBoxList(props) {
	const feedBoxes = props.feeds.map((feed) =>
  	<FeedBox title={feed.title} url={feed.url} resource={feed.resource} limit={feed.itemLimit} key={feed.title} />
	);

	return (
	  <ul className="drag-item-list" id={props.id}>
	    {feedBoxes}
	  </ul>
  );
}

class FeedBox extends React.Component {
	  constructor(props) {
	    super(props);
	    this.state = {
				items: [],
				title: props.title,
				url: props.url,
				resource: props.resource,
				limit: props.limit
			};
	  }

	  componentDidMount() {
	    // runs after the component output has been rendered to the DOM
			this.__fetchRSSFeedItems();
			this.timerID = setInterval(() => this.__fetchRSSFeedItems(), 5*60*1000);
	  }

		componentDidUpdate() {
		}

	  componentWillUnmount() {
	    // runs before the component is removed from the DOM
			clearInterval(this.timerID);
	  }

		__fetchRSSFeedItems() {
				//console.log("Fetching", this.state.title, "items @", this.state.resource, '...');
				fetch(this.state.resource + "?limit=" + this.state.limit)
					.then(response => response.json())
					.then(respData => {
						const items = respData.Items;

						this.setState(function(state, props) {
							return { items: items };
						});

					});
		}

	  render() {
			let feedItems = [];
			if ( this.state.items ) {
			  feedItems = this.state.items.map((item) =>
			    <FeedItem url={item.Url} title={item.Title} description={item.Description} published={item.Published} key={item.Url} />
			  );
			}
		  return (
		    <li className="drag-item">
		      <FeedBoxHeader text={this.state.title} url={this.state.url} />
		      <ul>
		        {feedItems}
		      </ul>
		    </li>
		  );
		}
	}

function FeedBoxHeader(props) {
	//<h2><a href={props.url} target="_blank">{props.text}</a></h2>
  return (
		<span className="drag-item-header">
			<h2>{props.text}</h2>
		</span>
	);
}

class FeedItem extends React.Component {
	constructor(props) {
		super(props);
		this.state = {
			url: props.url,
			title: props.title,
			description: props.description,
			published: new Date(props.published)
		}
	}

	__getClass() {
		let now = new Date();
		let age = Math.abs(now - this.state.published);
		let day_ms = 24*60*60*1000;
		let half_day_ms = day_ms * 0.5;

		if (day_ms <= age) {
			return "day_old";
		}

		if (half_day_ms <= age) {
			return "half_day_old";
		}

		return "";
	}

	render() {
		return (
			<li>
				<a className={this.__getClass()} href={this.state.url} target="_blank">{this.state.title}</a>
			</li>
		);
	}
}

class App extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
			feedColumns: []
		};
  }

  componentDidMount() {
    // runs after the component output has been rendered to the DOM
		this.__fetchRSSFeeds();
		this.timerID = setInterval(() => this.__fetchRSSFeeds(), 5*60*1000);
  }

	componentDidUpdate() {
		this.__enableDragAndDrop();
	}

  componentWillUnmount() {
    // runs before the component is removed from the DOM
		clearInterval(this.timerID);
  }

	__enableDragAndDrop() {
		var feedLists = document.querySelectorAll('.drag-item-list');

		reactDragula(Array.from(feedLists))
		.on('drag', function(el) {
			el.classList.add('is-moving');
		})
		.on('dragend', function(el) {
			el.classList.remove('is-moving');

			// add the 'is-moved' class for 600ms then remove it
			window.setTimeout(function() {
				el.classList.add('is-moved');
				window.setTimeout(function() {
					el.classList.remove('is-moved');
				}, 200);
			}, 100);
		});
	}

	__getFeedIdx(arr, feedName) {
		for (var i = 0; i < arr.length; i++) {
			if (arr[i].title === feedName ) {
				return i;
			}
		}
		return -1;
	}

	__initFeedColumn(arr, columnIdx) {
		if ( arr[columnIdx] === undefined ) {
			let colors = ["red", "blue", "green"]; // TODO: make it configurable
			arr[columnIdx] = {color: colors[columnIdx] || "yellow", feeds: []};
		}
	}

	__markUnusedFeedsForRemoval(feedColumns) {
		for (var feedColIdx = 0; feedColIdx < feedColumns.length; feedColIdx++) {
			if ( 'feeds' in feedColumns[feedColIdx] ) {
				for (var feedIdx = 0; feedIdx < feedColumns[feedColIdx].feeds.length; feedIdx++) {
					feedColumns[feedColIdx].feeds[feedIdx].canRemove = true;
				}
			}
		}
	}

	__removeUnusedFeeds(feedColumns) {
		for (var feedColIdx = 0; feedColIdx < feedColumns.length; feedColIdx++) {
			if ( 'feeds' in feedColumns[feedColIdx] ) {
				feedColumns[feedColIdx].feeds = feedColumns[feedColIdx].feeds.filter(function(element, index, arr) {
					return (element.canRemove === false);
				});
			}
		}
	}

	__fetchRSSFeeds() {
			//console.log("Fetching feeds...");
			fetch(myConfig.API_URI)
				.then(response => response.json())
				.then(respData => {
					const feeds = respData.Feeds;

					this.setState(function(state, props) {
						this.__markUnusedFeedsForRemoval(state.feedColumns);

						for ( var i = 0; i < feeds.length; i++ ) {
							let name 		 = feeds[i].Name;
							let col  		 = feeds[i].Column;
							let url  		 = feeds[i].Url;
							let resource = feeds[i].Resource;
							let itemLimit = feeds[i].ItemLimit;

							this.__initFeedColumn(state.feedColumns, col);

							let feedIdx = this.__getFeedIdx(state.feedColumns[col].feeds, name);
							if (0 <= feedIdx) {
								state.feedColumns[col].feeds[feedIdx].canRemove = false;
							} else {
								let feed = {
									title: name,
									url: url,
									resource: resource,
									itemLimit: itemLimit,
									canRemove: false
								};
								state.feedColumns[col].feeds.push(feed);
							}
						}

						this.__removeUnusedFeeds(state.feedColumns);

						return { feedColumns: state.feedColumns };

					});
				});
	}

  render() {
    const columns = this.state.feedColumns.map((column, index) =>
      <Column id={index} color={column.color} feeds={column.feeds} key={index}/>
    );

    return (
      <ul id="columns">
        {columns}
      </ul>
    );
  }
}

ReactDOM.render(<App />, document.getElementById('root'));
